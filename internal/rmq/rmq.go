package rmq

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

var ErrStopReconn = errors.New("stop reconnecting")

type Rmq struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	connClosed chan struct{}

	uri          string
	exchangeName string
	exchangeType string
	queueName    string
	bindingKey   string

	reConnMaxElapsedTime  time.Duration
	reConnInitialInterval time.Duration
	reConnMultiplier      float64
	reConnMaxInterval     time.Duration
}

func New(
	uri, exchangeName, exchangeType, queueName, bindingKey, reConnMaxElapsedTime, reConnInitialInterval string,
	reConnMultiplier float64,
	reConnMaxInterval string,
) (*Rmq, error) {
	reConnMaxElapsedTimeDur, err := time.ParseDuration(reConnMaxElapsedTime)
	if err != nil {
		return nil, errors.Wrapf(err, "reconnection max elapsed time parsing fail (%s)", reConnMaxElapsedTime)
	}

	reConnInitialIntervalDur, err := time.ParseDuration(reConnInitialInterval)
	if err != nil {
		return nil, errors.Wrapf(err, "reconnection initial interval parsing fail (%s)", reConnInitialInterval)
	}

	reConnMaxIntervalDur, err := time.ParseDuration(reConnMaxInterval)
	if err != nil {
		return nil, errors.Wrapf(err, "reconnection max interval parsing fail (%s)", reConnMaxInterval)
	}

	return &Rmq{
		uri:          uri,
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		queueName:    queueName,
		bindingKey:   bindingKey,

		reConnMaxElapsedTime:  reConnMaxElapsedTimeDur,
		reConnInitialInterval: reConnInitialIntervalDur,
		reConnMultiplier:      reConnMultiplier,
		reConnMaxInterval:     reConnMaxIntervalDur,
	}, nil
}

func (r *Rmq) Init(ctx context.Context) error {
	var err error
	startTime := time.Now()

	for time.Since(startTime) < 30*time.Second {
		if err = r.connect(ctx); err == nil {
			// Connection success, then exit from loop
			break
		}
		// If the connection fails, wait a little before trying again.
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return errors.Wrap(err, "rmq connection fail")
	}

	if err := r.prepareQueue(); err != nil {
		return errors.Wrap(err, "preparing queue fail")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.connClosed:
				if err := r.reConnect(ctx); err != nil {
					log.Fatal().Err(err).Msg("reconnecting error")
				}
			}
		}
	}()

	return nil
}

func (r *Rmq) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}

func (r *Rmq) Publish(msg amqp.Publishing) error {
	ctx := context.Background()
	if r.channel == nil {
		return nil
	}
	if err := r.channel.PublishWithContext(ctx, r.exchangeName, r.queueName, false, false, msg); err != nil {
		return errors.Wrap(err, "rmq publish fail")
	}

	return nil
}

func (r *Rmq) Consume(consumerTag string) (<-chan amqp.Delivery, error) {
	msgsCh, err := r.channel.Consume(
		r.queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "rmq consume fail")
	}

	return msgsCh, nil
}

func (r *Rmq) Qos(prefetchCount int) error {
	if err := r.channel.Qos(prefetchCount, 0, false); err != nil {
		return errors.Wrap(err, "error setting qos")
	}

	return nil
}

func (r *Rmq) IsClosed() bool {
	return r.conn.IsClosed()
}

// Reconnecting algo.
func (r *Rmq) reConnect(ctx context.Context) error {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = r.reConnMaxElapsedTime
	be.InitialInterval = r.reConnInitialInterval
	be.Multiplier = r.reConnMultiplier
	be.MaxInterval = r.reConnMaxInterval

	b := backoff.WithContext(be, ctx)
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return ErrStopReconn
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(d):
			log.Warn().Str("after", d.String()).Msg("reconnection")

			if err := r.connect(ctx); err != nil {
				log.Error().Err(err).Msg("couldn't connect in reconnect call")
				continue
			}
			if err := r.prepareQueue(); err != nil {
				log.Error().Err(err).Msg("couldn't preparing queue in reconnect call")
				continue
			}

			return nil
		}
	}
}

// Connect to RabbitMQ.
func (r *Rmq) connect(ctx context.Context) error {
	var err error

	r.conn, err = amqp.Dial(r.uri)
	if err != nil {
		return errors.Wrap(err, "dial fail")
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel fail")
	}

	r.connClosed = make(chan struct{})

	// Event for closing channel
	go func() {
		select {
		case <-ctx.Done():
		case <-r.conn.NotifyClose(make(chan *amqp.Error)):
			close(r.connClosed)
		}
	}()

	if err := r.channel.ExchangeDeclare(
		r.exchangeName,
		r.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "exchange declare fail")
	}

	return nil
}

// Declare queue.
func (r *Rmq) prepareQueue() error {
	_, err := r.channel.QueueDeclare(
		r.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "queue declare fail")
	}

	// Create a binding (exchange rule).
	if err = r.channel.QueueBind(
		r.queueName,
		r.bindingKey,
		r.exchangeName,
		false,
		nil,
	); err != nil {
		return errors.Wrap(err, "queue bind fail")
	}

	return nil
}
