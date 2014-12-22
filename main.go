package main

import (
	"github.com/docopt/docopt-go"
	"strconv"
	"fmt"
)

var version string = "0.0.0"

type Config struct {
	Endpoint         string
    VolumeDir        string
    InspectFrequency int
}

func usage() string {
	return `Testradamus.

Usage:
	damus init <app> [-d <endpoint> | --docker-host=<endpoint>]
	damus test <app> [-f <ms> | --freq=<ms>] [-v <dir> | --volume=<dir>] [-d <endpoint> | --docker-host=<endpoint>]
	damus flush <app> [-b <name> | --build=<name>] [-d <endpoint> | --docker-host=<endpoint>]
	damus --help
	damus --version

Options:
	-h --help                               Show this screen.
	--version                               Show version.
	-f <ms> --freq=<ms>                     Inspection frequency in milliseconds [default: 1000].
	-v <dir> --volume=<dir>                 Pass a volume to the containers [default: ./logs].
	-b <name> --build=<name>                Flush up to and including the specified build.
	-d <endpoint> --docker-host=<endpoint>  Specify a different Docker host [default: unix:///var/run/docker.sock].`
}

func args() (map[string]interface{}, error) {
	v := fmt.Sprintf("Testradamus %s", version)
	args, err := docopt.Parse(usage(), nil, true, v, false)
	if err != nil {
		return nil, err
	}

	freq, err := strconv.Atoi(args["--freq"].(string))
	if err != nil {
		return nil, err
	}
	args["freq"] = freq

	return args, nil
}

func conf(args map[string]interface{}) (string, Config) {
	name := args["<app>"].(string)

	c := Config{}
	if args["--docker-host"] != nil {
		c.Endpoint = args["--docker-host"].(string)
	}

	if args["--volume"] != nil {
		c.VolumeDir = args["--volume"].(string)
	}

	if args["freq"] != nil {
		c.InspectFrequency = args["freq"].(int)
	}

	return name, c
}

func flushTags(args map[string]interface{}) ([]string) {
	if args["--base"].(bool) {
		return []string{"base", "app", "test"}
	}

	if args["--app"].(bool) {
		return []string{"app", "test"}
	}

	return []string{"test"}
}

func main() {
	args, err := args()
	if err != nil {
		panic(err)
	}

	a, err := NewApp(conf(args))
	if err != nil {
		panic(err)
	}

	if args["init"].(bool) {
		a.Init()
	} else if args["test"].(bool) {
		a.Test()
	} else if args["flush"].(bool) {
		var chain Chain
		if args["--build"] != nil {
			chain = a.Builds.Get(args["--build"].(string))
		} else {
			chain = a.Builds.Final()
		}
		a.Flush(chain)
	}
}
