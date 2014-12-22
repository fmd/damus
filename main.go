package main
/*
import (
	"os"
	"fmt"
	"time"
	"hash/fnv"
	"encoding/hex"
	"github.com/fsouza/go-dockerclient"
)
*/

/*
var client *docker.Client
var image string = 'xxx-yyy'
var timehash string
var builder string
var cukes string
var rspec string
var canor string

func cleanup() {
	fmt.Printf("Cleaning up... ")
	kill(builder)
	kill(cukes)
	kill(rspec)
	kill(canor)
	fmt.Println("Done.")
}

func kill(id string) {
	err := client.StopContainer(id, 15)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func inspect(id string) *docker.Container {
	c, err := client.InspectContainer(id)
	if err != nil {
		cleanup()
		panic(err)
	}
	return c
}

func timeHash() string {
	h := fnv.New64a()
	defer h.Reset()
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}

func name(id string) string {
	return fmt.Sprintf("%s-%s",id,timehash)
}

func commit (id string) {
	fmt.Printf("Committing changes... ")
	opts := docker.CommitContainerOptions{
		Container: id,
		Repository: image,
		Tag: "latest",
	}

	_, err := client.CommitContainer(opts)
	if err != nil {
		cleanup()
		panic(err)
	}
	fmt.Println("Done.")
}

func build() {
	builder = create(name("builder"), "/app/quick-build")
	start(builder)
	fmt.Printf("Rebuilding container... ")
	code := func() int {
		for {
			state := inspect(builder).State
			if !state.Running {
				return state.ExitCode
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	fmt.Printf("Finished with code %d.\n", code)
	if code == 4 {
		commit(builder)
	}
}
*/

/*
	endpoint := "unix:///var/run/docker.sock"
	client,_ = docker.NewClient(endpoint)

	timehash = timeHash()

	build()

	cukes = create(name("cucumber"), "/app/cucumber")
	canor = create(name("canorman"), "/app/canorman")
	rspec = create(name("rspec"), "/app/rspec")

	start(cukes)
	start(canor)
	start(rspec)

	fmt.Printf("Building...")
	for {
		specState := inspect(rspec).State
		canorState := inspect(canor).State
		cukesState := inspect(cukes).State
		fmt.Printf(".")

		if specState.ExitCode != 0 || canorState.ExitCode != 0 || cukesState.ExitCode != 0 {
			fmt.Println("\nBuild Failed!")
			cleanup()
			fmt.Println("\nExiting with code 1.")
			os.Exit(1)
		}

		if !specState.Running && !canorState.Running && !cukesState.Running {
			fmt.Println("\nBuild Successful!")
			cleanup()
			fmt.Println("\nExiting with code 0.")
			os.Exit(0)
		}
		time.Sleep(3000 * time.Millisecond)
	}
	*/

import (
	"github.com/docopt/docopt-go"
	"strconv"
	"errors"
	"fmt"
	"os"
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
	damus test <app> [-f <ms> | --freq=<ms>] [-v <dir> | --volume=<dir>]
	damus flush <app> [-b <name> | --build=<name>]
	damus --help
	damus --version

Options:
	-h --help                          Show this screen.
	--version                          Show version.
	-f <ms> --freq=<ms>                Inspection frequency in milliseconds [default: 1000].
	-v <dir> --volume=<dir>            Pass a volume to the containers [default: ./logs].
	-b <name> --build=<name>           Flush up to and including the specified build.
	-d <endpoint> --docker=<endpoint>  Specify a different Docker host.`
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

	if args["--docker"] != nil {
		c.Endpoint = args["--docker"].(string)
	} else {
		c.Endpoint = "unix:///var/run/docker.sock"
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
		err = a.Init()
	} else if args["test"].(bool) {
		code, err := a.Test()
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(code)
	} else if args["flush"].(bool) {
		var chain Chain
		if args["--build"] != nil {
			chain, err = a.Builds.Get(args["--build"].(string))
		} else {
			chain = a.Builds.Final()
		}

		if err == nil {
			err = a.Flush(chain)
		}
	} else {
		err = errors.New("Bad command!")
	}

	if err != nil {
		panic(err)
	}
}
