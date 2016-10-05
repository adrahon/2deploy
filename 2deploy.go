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

    // # Get stack name from --name
	// # Get stack name from directory if not passed 
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

    // Networks

    if project.NetworkConfigs == nil || len(project.NetworkConfigs) == 0 {
        // if no network create default
        fmt.Println("No networks!")
    } else {
        for name, config := range project.NetworkConfigs {
            // # if network external check if exists
            if config.External.External {
                fmt.Println(fmt.Sprintf("Network: %s (external)", name))
                // handle external name
                if config.External.Name != "" {
                    fmt.Println(fmt.Sprintf("Network: %s (external: %s)", name, config.External.Name))
                }
            } else {
                // # else create network
                // # if no driver set default
                if config.Driver != "" {
                    fmt.Println(fmt.Sprintf("Network: %s (driver: %s)", name, config.Driver))
                } else {
                    fmt.Println(fmt.Sprintf("Network: %s (driver: default)", name))
                }
            }
        }
    }

    // # Volumes

	// # Services
    // # Dependencies?
   
    // # Timeouts / Errors

}
