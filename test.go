package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
)

type Test struct {
	Cmd  []string `json:"cmd"`
	Name string   `json:"name"`
	Id   string   `json:"-"`
}

func (t Test) Config(image string, build string, stamp string) docker.CreateContainerOptions {
	opts := &docker.Config{
		Image:        fmt.Sprintf("%s-%s", image, build),
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          t.Cmd,
	}
	return docker.CreateContainerOptions{Name: fmt.Sprintf("%s-%s", t.Name, stamp), Config: opts}
}
