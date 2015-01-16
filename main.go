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
    NoCache          bool
    Quiet            bool
}

func usage() string {
	return `Testradamus.

Usage:
	damus init <app> [-q | --quiet] [-n | --no-cache] [-d <endpoint> | --docker-host=<endpoint>]
	damus test <app> [-q | --quiet] [-f <ms> | --freq=<ms>] [-v <dir> | --volume=<dir>] [-d <endpoint> | --docker-host=<endpoint>]
	damus flush <app> [-b <name> | --build=<name>] [-d <endpoint> | --docker-host=<endpoint>]
	damus --help
	damus --version

Options:
	-h --help                               Show this screen.
	--version                               Show version.
	-q --quiet                              Suppresses build output [default: false].
	-n --no-cache                           Suppresses build output [default: false].
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

	if args["--quiet"] != nil {
		c.Quiet = args["--quiet"].(bool)
	}

	if args["--no-cache"] != nil {
		c.NoCache = args["--no-cache"].(bool)
	}

	if args["--volume"] != nil {
		c.VolumeDir = args["--volume"].(string)
	}

	if args["freq"] != nil {
		c.InspectFrequency = args["freq"].(int)
	}

	return name, c
}

func main() {
	args, err := args()
	if err != nil {
		panic(err)
	}

	a := NewApp(conf(args))

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
