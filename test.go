package main

import (
    "github.com/fsouza/go-dockerclient"
    "fmt"
)

type Test struct {
    Cmd       string            `json:"cmd"`
    Name      string            `json:"name"`
    Container *docker.Container `json:"-"`
}

func (t Test) Config(image string, build string) docker.CreateContainerOptions {
    opts := &docker.Config{
        Image: fmt.Sprintf("%s-%s", image, build),
        Tty: true,
        AttachStdout: true,
        AttachStderr: true,
        Cmd: []string{t.Cmd},
    }
    return docker.CreateContainerOptions{Name: t.Name, Config: opts}
}
