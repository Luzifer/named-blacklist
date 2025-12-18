package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/generator"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Config         string `flag:"config" default:"config.yaml" description:"Config file to use for generating the file"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	conf *config.File

	version = "dev"
)

func initApp() (err error) {
	rconfig.AutoEnv(true)
	if err = rconfig.ParseAndValidate(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	return nil
}

func main() {
	var err error

	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("named-blacklist %s\n", version) //nolint:forbidigo
		os.Exit(0)
	}

	if conf, err = config.LoadConfigFile(cfg.Config); err != nil {
		logrus.WithError(err).Fatal("reading config file")
	}

	blacklist, err := generator.GenerateBlacklist(version, conf.Providers)
	if err != nil {
		logrus.WithError(err).Fatal("generating blacklist")
	}

	if err = conf.CompiledTemplate.Execute(os.Stdout, map[string]any{
		"blacklist": blacklist,
	}); err != nil {
		logrus.WithError(err).Fatal("rendering blacklist")
	}
}
