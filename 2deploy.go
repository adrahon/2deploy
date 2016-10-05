package main

import (
    "log"
    "fmt"
    "os"
    "path"
    "context"

    "github.com/docker/libcompose/config"
    "github.com/docker/libcompose/project"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
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

    if project.VolumeConfigs == nil || len(project.VolumeConfigs) == 0 {
        // no volumes
        fmt.Println("No volumes")
    } else {
        for name, config := range project.VolumeConfigs {
            // # if volume external check if exists
            if config.External.External {
                fmt.Println(fmt.Sprintf("Volume: %s (external)", name))
                // handle external name
                if config.External.Name != "" {
                    fmt.Println(fmt.Sprintf("Volume: %s (external: %s)", name, config.External.Name))
                }
            } else {
                // # else create volume
                fmt.Println(fmt.Sprintf("Volume: %s", name))
            }
        }
    }

	// # Services

    if project.ServiceConfigs == nil {
        // no service, abort?
        fmt.Println("No services")
    } else {
        for name, config := range project.ServiceConfigs.All() {
            // image, ports, networks, volumes
            fmt.Println(fmt.Sprintf("Service: %s", project.Name + "_" + name))
            if config.Image != "" {
                fmt.Println(fmt.Sprintf("  Image: %s", config.Image))
            } else {
                // # if no image abort
                fmt.Println("  No image!")
            }
            for _, port := range config.Ports {
                fmt.Println(fmt.Sprintf("  Port: %s", port))
            }
            if config.Networks != nil && len(config.Networks.Networks) != 0 {
                for _, network := range config.Networks.Networks {
                    fmt.Println(fmt.Sprintf("  Network: %s", network.RealName))
                }
            }
            if config.Volumes != nil && len(config.Volumes.Volumes) != 0 {
                for _, volume := range config.Volumes.Volumes {
                    fmt.Println(fmt.Sprintf("  Volume: %s", volume))
                }
            }
        }

		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}

		options := types.ImageListOptions{All: true}
		images, err := cli.ImageList(context.Background(), options)
		if err != nil {
			panic(err)
		}

		for _, c := range images {
			fmt.Println(c.ID)
		}
	}

	// # Exposed Ports
	// # Dependencies?
	// # More services config params 
	// # Timeouts / Errors

}
