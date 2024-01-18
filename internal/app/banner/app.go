package banner

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cronnoss/banners-rotation/interfaces"
	"github.com/cronnoss/banners-rotation/internal/config"
	"github.com/cronnoss/banners-rotation/internal/logger"
	"github.com/cronnoss/banners-rotation/internal/rmq"
	internalgrpc "github.com/cronnoss/banners-rotation/internal/server/grpc"
	"github.com/cronnoss/banners-rotation/internal/server/pb"
	"github.com/cronnoss/banners-rotation/internal/storage/sql"
	"google.golang.org/grpc"
)

type App struct {
	logger     interfaces.Logger
	storage    interfaces.Storage
	serverGRPC *grpc.Server
}

func NewApp(ctx context.Context, conf *config.BannerConfig) (*App, error) {
	app := &App{}

	logger := logger.New(conf.Logger.Level, os.Stdout)
	app.logger = logger

	// Initializing the data store.
	psqlStorage := new(sql.Storage)
	if err := psqlStorage.Connect(
		ctx,
		conf.Database.Port,
		conf.Database.Host,
		conf.Database.Username,
		conf.Database.Password,
		conf.Database.Dbname,
	); err != nil {
		return nil, fmt.Errorf("cannot connect to PostgreSQL: %w", err)
	}
	err := psqlStorage.Migrate(ctx, conf.Storage.Migration)
	if err != nil {
		return nil, fmt.Errorf("migration did not work out: %w", err)
	}
	app.storage = psqlStorage

	// Initializing RMQ.
	URI := fmt.Sprintf("%s://%s:%s@%s:%d/", // "amqp://guest:guest@localhost:5672/"
		conf.RMQ.RabbitmqProtocol,
		conf.RMQ.RabbitmqUsername,
		conf.RMQ.RabbitmqPassword,
		conf.RMQ.RabbitmqHost,
		conf.RMQ.RabbitmqPort,
	)
	logger.Info("URI: %s", URI)

	eventsProdMq, err := rmq.New(
		URI,
		conf.Queues.Events.ExchangeName,
		conf.Queues.Events.ExchangeType,
		conf.Queues.Events.QueueName,
		conf.Queues.Events.BindingKey,
		conf.RMQ.ReConnect.MaxElapsedTime,
		conf.RMQ.ReConnect.InitialInterval,
		conf.RMQ.ReConnect.Multiplier,
		conf.RMQ.ReConnect.MaxInterval,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RMQ for scheduler: %w", err)
	}

	if err := eventsProdMq.Init(ctx); err != nil {
		logger.Error("RMQ initialization failed: %v", err)
	}

	// Initializing gRPC server.
	app.serverGRPC = grpc.NewServer(
		grpc.UnaryInterceptor(internalgrpc.NewLoggingInterceptor(logger).UnaryServerInterceptor),
	)

	api := internalgrpc.NewEventServiceServer(app.storage, eventsProdMq, logger)
	pb.RegisterBannerServiceServer(app.serverGRPC, api)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.GRPC.Host, conf.GRPC.Port))
	if err != nil {
		logger.Error("Failed to listen: %v", err)
	}

	go func() {
		logger.Info("Starting gRPC server on port %s", fmt.Sprintf("%s:%d", conf.GRPC.Host, conf.GRPC.Port))
		if err := app.serverGRPC.Serve(grpcListener); err != nil {
			logger.Error("gRPC server failed: %v", err)
		}
	}()

	// Waiting for the signal to stop the gRPC server.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		// A completion signal has been received. Gracefully stop the gRPC server.
		app.serverGRPC.GracefulStop()
		logger.Info("gRPC server stopped")
	}()

	return app, nil
}
