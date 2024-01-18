package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Configure interface {
	Init(file string) error
}

type LoggerConf struct {
	Level       string `json:"loggerLevel"`
	Development bool   `json:"loggerDevelopment"`
}

type StorageConf struct {
	Migration string `json:"migration"`
}

type DataBaseConf struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GRPC struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type RMQ struct {
	RabbitmqProtocol string `json:"rabbitmqProtocol"`
	RabbitmqUsername string `json:"rabbitmqUsername"`
	RabbitmqPassword string `json:"rabbitmqPassword"`
	RabbitmqHost     string `json:"rabbitmqHost"`
	RabbitmqPort     int    `json:"rabbitmqPort"`
	ReConnect        struct {
		MaxElapsedTime  string  `json:"maxElapsedTime"`
		InitialInterval string  `json:"initialInterval"`
		Multiplier      float64 `json:"multiplier"`
		MaxInterval     string  `json:"maxInterval"`
	}
}

type Queue struct {
	ExchangeName string `json:"exchangeName"`
	ExchangeType string `json:"exchangeType"`
	QueueName    string `json:"queueName"`
	BindingKey   string `json:"bindingKey"` // This is the message routing rule.
}

type Consumer struct {
	ConsumerTag      string  `json:"consumerTag"`
	QosPrefetchCount float64 `json:"qosPrefetchCount"`
	Threads          float64 `json:"threads"`
}

func Init(file string, c Configure) (Configure, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigFile(file)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "read config file failed")
	}

	if err := viper.Unmarshal(c); err != nil {
		return nil, errors.Wrap(err, "unmarshal config file failed")
	}

	return c, nil
}
