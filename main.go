package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Config         string `flag:"config" default:"config.yaml" description:"Config file to use for generating the file"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	config *configfile

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
	var (
		blacklist []entry
		whitelist []entry
		write     = new(sync.Mutex)
		wg        sync.WaitGroup
		err       error
	)

	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("named-blacklist %s\n", version) //nolint:forbidigo
		os.Exit(0)
	}

	if config, err = loadConfigFile(cfg.Config); err != nil {
		logrus.WithError(err).Fatal("reading config file")
	}

	wg.Add(len(config.Providers))
	for _, p := range config.Providers {
		go func(p providerDefinition) {
			defer wg.Done()

			logger := logrus.WithField("provider", p.Name)
			logger.Info("starting domain list extraction")

			entries, err := getDomainList(p)
			if err != nil {
				logger.WithError(err).Fatal("getting domain list")
			}

			write.Lock()
			defer write.Unlock()

			for _, e := range entries {
				switch p.Action {
				case providerActionBlacklist:
					blacklist = addIfNotExists(blacklist, e)

				case providerActionWhitelist:
					whitelist = addIfNotExists(whitelist, e)

				default:
					logger.Fatalf("Inavlid action %q", p.Action)
				}
			}

			logger.WithField("no_entries", len(entries)).Info("extraction complete")
		}(p)
	}

	wg.Wait()

	blacklist = cleanFromList(blacklist, whitelist)

	sort.Slice(blacklist, func(i, j int) bool { return blacklist[i].Domain < blacklist[j].Domain })

	if err = config.tpl.Execute(os.Stdout, map[string]any{
		"blacklist": blacklist,
	}); err != nil {
		logrus.WithError(err).Fatal("rendering blacklist")
	}
}

func addIfNotExists(entries []entry, e entry) []entry {
	for i, pe := range entries {
		if pe.Domain == e.Domain {
			entries[i].Comments = append(pe.Comments, e.Comments...) //nolint:gocritic // This accumulates comments on an existing entry
			return entries
		}
	}

	return append(entries, e)
}

func cleanFromList(blacklist, whitelist []entry) []entry {
	var tmp []entry

	for _, be := range blacklist {
		var found bool

		for _, we := range whitelist {
			if we.Domain == be.Domain {
				found = true
				break
			}
		}

		if !found {
			tmp = append(tmp, be)
		}
	}

	return tmp
}
