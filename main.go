package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

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

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("named-blacklist %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	var (
		blacklist []entry
		whitelist []entry
		write     = new(sync.Mutex)
		wg        sync.WaitGroup
		err       error
	)

	if config, err = loadConfigFile(cfg.Config); err != nil {
		log.WithError(err).Fatal("Unable to read config file")
	}

	wg.Add(len(config.Providers))
	for _, p := range config.Providers {

		go func(p providerDefinition) {
			defer wg.Done()

			logger := log.WithField("provider", p.Name)
			logger.Info("Starting domain list extraction")

			entries, err := getDomainList(p)
			if err != nil {
				logger.
					WithError(err).
					Fatal("Unable to get domain list")
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

			logger.WithField("no_entries", len(entries)).Info("Extraction complete")
		}(p)

	}

	wg.Wait()

	blacklist = cleanFromList(blacklist, whitelist)

	sort.Slice(blacklist, func(i, j int) bool { return blacklist[i].Domain < blacklist[j].Domain })

	config.tpl.Execute(os.Stdout, map[string]interface{}{
		"blacklist": blacklist,
	})
}

func addIfNotExists(entries []entry, e entry) []entry {
	var (
		found bool
		out   []entry
	)

	for _, pe := range entries {
		if pe.Domain == e.Domain {
			found = true
			pe.Comment = strings.Join([]string{pe.Comment, e.Comment}, ", ")
		}

		out = append(out, pe)
	}

	if !found {
		out = append(out, e)
	}

	return out
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
