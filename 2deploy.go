package main

import (
    "log"
    "fmt"
    "os"
    "path"

    "golang.org/x/net/context"

    "github.com/docker/libcompose/docker"
    "github.com/docker/libcompose/docker/ctx"
    "github.com/docker/libcompose/project"
    "github.com/docker/libcompose/project/options"
)

func main() {
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    _, dir := path.Split(pwd)

    project, err := docker.NewProject(&ctx.Context{
        Context: project.Context{
            ComposeFiles: []string{"docker-compose.yml"},
            ProjectName:  dir,
        },
    }, nil)

    if err != nil {
        log.Fatal(err)
    }

    err = project.Up(context.Background(), options.Up{})

    if err != nil {
       log.Fatal(err)
    }
}
