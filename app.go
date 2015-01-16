package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

var AppJson string = "app.json"
var AppDir string = "apps"
var LogDir string = "logs"

type App struct {
	Committer Test   `json:"committer"`
	Tests     []Test `json:"testers"`
	Builds    Chain  `json:"builds"`

	Name   string `json:"-"`
	Config Config `json:"-"`
}

func (a *App) Build(builds Chain) {
	b, err := NewBuilder(a.Config.Endpoint, a.Name, a.Config.NoCache, a.Config.Quiet)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	for _, build := range a.Builds {
		err = b.Build(build)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	}
}

func (a *App) Init() {
	a.Flush(a.Builds)
	a.Build(a.Builds)
}

func (a *App) Fix() {
	b, err := NewBuilder(a.Config.Endpoint, a.Name, a.Config.NoCache, a.Config.Quiet)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	for _, build := range a.Builds {
		exists, err := b.ImageExists(fmt.Sprintf("%s-%s", a.Name, build))
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}

		if !exists {
			err = b.Build(build)
			if err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
		}
	}
}

func (a *App) Flush(builds Chain) {
	b, err := NewBuilder(a.Config.Endpoint, a.Name, a.Config.NoCache, a.Config.Quiet)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	for _, build := range builds {
		err = b.Remove(build)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	}
}

func (a *App) Test() {
	a.Fix()

	//Create builder
	b, err := NewBuilder(a.Config.Endpoint, a.Name, a.Config.NoCache, a.Config.Quiet)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	//Ensure testable build
	final := a.Builds.Final().String()
	exists, err := b.ImageExists(fmt.Sprintf("%s-%s", a.Name, final))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if !exists {
		st := fmt.Sprintf("Image '%s-%s' doesn't exist. Try running `damus init` first.", a.Name, final)
		log.Error(st)
		os.Exit(1)
	}

	//Create tester
	t, err := NewTester(a.Config.Endpoint, a.Name, final, LogDir, a.Config.InspectFrequency)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	//Run commit step
	err = t.Commit(&a.Committer)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	//Test the tests
	code, err := t.Test(a.Tests)
	if err != nil {
		log.Error(err.Error())
	}

	os.Exit(code)
}

func (a *App) Parse() error {
	path, err := a.Path()
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadFile(filepath.Join(path, AppJson))
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, a)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Path() (string, error) {
	path := ""
	files, err := ioutil.ReadDir(AppDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		if file.Name() == a.Name {
			path = filepath.Join(AppDir, file.Name())
		}
	}

	if len(path) == 0 {
		err = errors.New(fmt.Sprintf("Could not find path for `%s`.", a.Name))
	}
	return path, err
}

func NewApp(name string, config Config) *App {
	a := &App{
		Name:   name,
		Config: config,
	}

	log.SetLevel(log.InfoLevel)
	err := a.Parse()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	return a
}
