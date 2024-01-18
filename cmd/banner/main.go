//go:generate protoc --go_out=../../internal/server/pb --proto_path=../../api/ ../../api/Service.proto
//go:generate protoc --go-grpc_out=../../internal/server/pb --proto_path=../../api/ ../../api/Service.proto
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cronnoss/banners-rotation/internal/app/banner"
	"github.com/cronnoss/banners-rotation/internal/config"
	"github.com/pkg/errors"
)

var bannerConfigFile string

func init() {
	flag.StringVar(&bannerConfigFile, "config", "banner_config.yaml", "Path to configuration file")
}

func main() {
	if err := mainImpl(); err != nil {
		log.Fatal(err)
	}
}

func mainImpl() error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	flag.Parse()

	if bannerConfigFile == "" {
		return fmt.Errorf("please set: '--config=<Path to configuration file>'")
	}

	conf := new(config.BannerConfig)
	if err := conf.Init(bannerConfigFile); err != nil {
		return errors.Wrap(err, "failed to init config")
	}

	_, err := banner.NewApp(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	return nil
}
