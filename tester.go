package main

import (
    "fmt"
    "time"
    "errors"
    "hash/fnv"
    "encoding/hex"
    "github.com/fsouza/go-dockerclient"
)

type Tester struct {
    Client           *docker.Client
    Image            string
    Build            string
    Stamp            string
    InspectFrequency int
}

func (t *Tester) FullName() string {
    return fmt.Sprintf("%s-%s", t.Image, t.Build)
}

func TimeHash() string {
    h := fnv.New64a()
    defer h.Reset()
    h.Write([]byte(time.Now().String()))
    return hex.EncodeToString(h.Sum(nil))
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
        Stamp: TimeHash(),
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
        Repository: t.FullName(),
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

type Result struct {
    Code int
    Error error
}

func (t *Tester) StartTest(test Test, results chan Result) {
    go func(test Test, results chan Result) {
        r := Result{
            Code: 0,
            Error: nil,
        }

        err := t.Create(&test)
        if err != nil {
            r.Code = 1
            r.Error = err
            results <- r
            return
        }

        err = t.Start(&test)
        if err != nil {
            r.Code = 1
            r.Error = err
            results <- r
            return
        }

        r.Code, r.Error = t.Check(&test)
        results <- r
        return
    }(test, results)
}

func (t *Tester) Test(s []Test) (int, error) {
    results := make(chan Result, len(s))

    for _, test := range s {
        t.StartTest(test, results)
    }

    finishedTests := 0

    for finishedTests != len(s) {
        select {
        case r := <-results:
            if r.Code != 0 {
                return r.Code, r.Error
            }
            if r.Error != nil {
                return r.Code, r.Error
            }
            finishedTests++
        }
    }

    return 0, nil
}

func (t *Tester) Create(s *Test) error {
    var err error
    s.Container, err = t.Client.CreateContainer(s.Config(t.Image, t.Build, t.Stamp))
    return err
}

func (t *Tester) Start(s *Test) error {
    if s.Container == nil {
        return errors.New("A container has not been created for this test.")
    }
    return t.Client.StartContainer(s.Container.ID, &docker.HostConfig{})
}
