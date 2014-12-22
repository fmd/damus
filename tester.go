package main

import (
    "fmt"
    "time"
    "errors"
    "github.com/fsouza/go-dockerclient"
)

type Tester struct {
    Client           *docker.Client
    Image            string
    Build            string
    InspectFrequency int
}

func NewTester(endpoint string, image string, build string, freq int) (*Tester, error) {
    c, err := docker.NewClient(endpoint)
    if err != nil {
        return nil, err
    }

    t := &Tester{
        Client: c,
        Image: image,
        Build: build,
        InspectFrequency: freq,
    }

    return t, nil
}

func (t *Tester) Check(s *Test) (int, error) {
    for {
        c, err := t.Client.InspectContainer(s.Container.ID)
        if err != nil {
            return 1, err
        }

        if c.State.ExitCode != 0 {
            return c.State.ExitCode, nil
        }

        if !c.State.Running {
            return 0, nil
        }

        time.Sleep(time.Duration(t.InspectFrequency) * time.Millisecond)
    }
}

func (t *Tester) SaveImage(s *Test) error {
    opts := docker.CommitContainerOptions{
        Container: s.Container.ID,
        Repository: s.Name,
        Tag: "latest",
    }
    _, err := t.Client.CommitContainer(opts)
    if err != nil {
        return err
    }
    return nil
}

func (t *Tester) Commit(s *Test) error {
    err := t.Create(s)
    if err != nil {
        return err
    }

    err = t.Start(s)
    if err != nil {
        return err
    }

    code, err := t.Check(s)
    if code != 0 {
        return errors.New("Commit action quit with non-zero exit code. Aborting.")
    }

    err = t.SaveImage(s)
    if err != nil {
        return err
    }

    return nil
}

func (t *Tester) Test(s *Test) error {
    fmt.Println(s)
    return nil
}

func (t *Tester) Create(s *Test) error {
    var err error
    s.Container, err = t.Client.CreateContainer(s.Config(t.Image, t.Build))
    return err
}

func (t *Tester) Start(s *Test) error {
    if s.Container == nil {
        return errors.New("A container has not been created for this test.")
    }
    return t.Client.StartContainer(s.Container.ID, &docker.HostConfig{})
}
