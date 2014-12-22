package main

import (
    "encoding/json"
    "path/filepath"
    "io/ioutil"
    "errors"
    "fmt"
)

var AppJson string = "app.json"
var AppDir  string = "apps"

type App struct {
    Committer Test      `json:"committer"`
    Tests     []Test    `json:"testers"`
    Builds    Chain     `json:"builds"`

    Name      string    `json:"-"`
    Config    Config    `json:"-"`
}

func (a *App) Init() error {
    var err error
    err = a.Flush(a.Builds)
    if err != nil {
        return err
    }
    err = a.Build(a.Builds)
    if err != nil {
        return err
    }
    return nil
}

func (a *App) Build(builds Chain) error {
    b, err := NewBuilder(a.Config.Endpoint, a.Name)
    if err != nil {
        return err
    }
    for _, build := range a.Builds {
        err = b.Build(build)
        if err != nil {
            return err
        }
    }
    return nil
}

func (a *App) Flush(builds Chain) error {
    b, err := NewBuilder(a.Config.Endpoint, a.Name)
    if err != nil {
        return err
    }
    for _, build := range builds {
        err = b.Remove(build)
        if err != nil {
            return err
        }
    }
    return nil
}

func (a *App) Test() (int, error) {
    b, err := NewBuilder(a.Config.Endpoint, a.Name)
    if err != nil {
      return 1, err
    }

    final := a.Builds.Final().String()
    exists, err := b.ImageExists(fmt.Sprintf("%s-%s", a.Name, final))
    if err != nil {
        return 1, err
    }

    if !exists {
        st := fmt.Sprintf("Image '%s-%s' doesn't exist. Try running `damus init` first.", a.Name, final)
        return 1, errors.New(st)
    }

    t, err := NewTester(a.Config.Endpoint, a.Name, final, a.Config.InspectFrequency)
    if err != nil {
      return 1, err
    }

    err = t.Commit(&a.Committer)
    if err != nil {
        return 1, err
    }

    return t.Test(a.Tests)
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

func NewApp(name string, config Config) (*App, error) {
    a := &App{
        Name: name,
        Config: config,
    }
    return a, a.Parse()
}
