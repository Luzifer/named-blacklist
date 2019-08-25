package main

import (
	"fmt"
	"os"
	"sort"
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

			entries, err := getDomainList(p)
			if err != nil {
				log.WithField("provider", p.Name).
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
					log.WithField("provider", p.Name).Fatalf("Inavlid action %q", p.Action)

				}
			}
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
	for _, pe := range entries {
		if pe.Domain == e.Domain {
			// Entry already exists, skip
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
