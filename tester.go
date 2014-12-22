package main

import (
    "errors"
    "github.com/fsouza/go-dockerclient"
)

type Tester struct {
    Client *docker.Client
    Image string
    Build string
}

func NewTester(endpoint string, image string, build string) (*Tester, error) {
    c, err := docker.NewClient(endpoint)
    if err != nil {
        return nil, err
    }

    t := &Tester{
        Client: c,
        Image: image,
        Build: build,
    }

    return t, nil
}

func (t *Tester) Commit() error {
    return nil
}

func (t *Tester) Test() error {
    return nil
}

func (t *Tester) Create(s *Test) error {
    var err error
    s.Container, err = t.Client.CreateContainer(s.Config(t.Image, t.Build))
    return err
}

func (t *Tester) Start(s *Test) error {
    if s.Container == nil {
        return errors.New("A container has not been created for this Tester.")
    }
    return t.Client.StartContainer(s.Container.ID, &docker.HostConfig{})
}
