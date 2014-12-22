package main

import (
    "os"
    "fmt"
    "strings"
    "path/filepath"
    "github.com/fsouza/go-dockerclient"
)

type Builder struct {
    Client *docker.Client
    Name  string
}

func (b *Builder) ImageExists(name string) (bool, error) {
    images, err := b.Client.ListImages(docker.ListImagesOptions{
        All: true,
        Filters: nil,
    })
    if err != nil {
        return false, err
    }
    fmt.Println(images)
    for _, image := range images {
        fmt.Println(image)
        fmt.Println(name)
        for _, t := range image.RepoTags {
            if strings.Split(t, ":")[0] == name {
                return true, nil
            }
        }
    }
    return false, nil
}

func NewBuilder(endpoint string, name string) (*Builder, error) {
    c, err := docker.NewClient(endpoint)
    if err != nil {
        return nil, err
    }
    b := &Builder{
        Client: c,
        Name: name,
    }
    return b, nil
}

func (b *Builder) BuildOptions(step string) docker.BuildImageOptions {
    return docker.BuildImageOptions{
        Name: fmt.Sprintf("%s-%s", b.Name, step),
        ForceRmTmpContainer: true,
        OutputStream: os.Stdout,
        ContextDir: filepath.Join("apps", b.Name, step),
    }
}

func (b *Builder) Build(step string) error {
    err := b.Client.BuildImage(b.BuildOptions(step))
    if err != nil {
        return err
    }
    return nil
}

func (b *Builder) Remove(step string) error {
    fullTag := fmt.Sprintf("%s-%s", b.Name, step)
    exists, err := b.ImageExists(fullTag)
    if err != nil || !exists {
        return err
    }
    return b.Client.RemoveImage(fullTag)
}