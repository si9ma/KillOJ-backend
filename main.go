package main

import (
	"os"

	"github.com/si9ma/KillOJ-backend/config"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/si9ma/KillOJ-common/log"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var (
	configPath = "conf/config.yml"
	app        *cli.App
	cfg        *config.Config
)

func init() {
	app = cli.NewApp()
	app.Name = "kbackend"
	app.Usage = "kbackend is a backend for KillOJ(https://github.com/si9ma/KillOJ)"
	app.Author = "si9ma"
	app.Email = "si9may@tom.com"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c",
			Value:       configPath,
			Destination: &configPath,
			Usage:       "Path to a configuration file",
		},
	}
	app.Action = func(ctx *cli.Context) (err error) {
		// Init configuration
		if cfg, err = Init(configPath); err != nil {
			log.Bg().Fatal("initialize fail", zap.String("configPath", configPath), zap.Error(err))
			return err
		}

		// setup Router
		r := setupRouter()
		if err := r.Run(cfg.App.Host); err != nil {
			log.Bg().Fatal("run backend fail", zap.Error(err))
			return err
		}

		return nil
	}

	// clean
	app.After = func(context *cli.Context) (err error) {
		// close db
		if gbl.DB != nil {
			if err = gbl.DB.Close(); err != nil {
				log.Bg().Error("close db fail", zap.Error(err))
			}
		}

		// close redis
		if gbl.Redis != nil {
			if err = gbl.Redis.Close(); err != nil {
				log.Bg().Error("close redis fail", zap.Error(err))
			}
		}

		// close tracer
		if err = gbl.TracerCloser.Close(); err != nil {
			log.Bg().Error("close tracer fail", zap.Error(err))
		}

		return err
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Bg().Fatal("run app fail", zap.Error(err))
	}
}
