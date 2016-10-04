package main

import (
    "log"
    "fmt"
    "os"
    "path"

    "github.com/docker/libcompose/config"
    "github.com/docker/libcompose/project"
)

func main() {
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    _, dir := path.Split(pwd)

    project := project.NewProject(&project.Context{
            ComposeFiles: []string{"docker-compose.yml"},
            ProjectName:  dir,
    }, nil, &config.ParseOptions{})

    if err := project.Parse(); err != nil {
        log.Fatal(err)
    }

    for name, _ := range project.NetworkConfigs {
        s := fmt.Sprintf("Network: %s", name)
        fmt.Println(s)
    }

}
