package main

import (
    "log"
    "fmt"
    "os"
    "path"
    "regexp"
    "strings"

    "golang.org/x/net/context"

    "github.com/docker/libcompose/config"
    "github.com/docker/libcompose/project"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
)

func main() {

    project_name := ProjectName()

    project := project.NewProject(&project.Context{
            ComposeFiles: []string{"docker-compose.yml"},
            ProjectName:  project_name,
    }, nil, &config.ParseOptions{})

    if err := project.Parse(); err != nil {
        log.Fatal(err)
    }

    
    cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    fmt.Println(fmt.Sprintf("cli: %s", cli.ClientVersion()))

    // # Check if stack exists

    // Networks

    if project.NetworkConfigs == nil || len(project.NetworkConfigs) == 0 {
        // if no network create default
        name := fmt.Sprintf("%s_default", project_name)
		config := config.NetworkConfig { Driver: "default", }
		err := NetworkCreate(cli, name, &config)
		if err != nil {
			fmt.Println(err)
		}
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
                // else create network
                realname := fmt.Sprintf("%s_%s", project_name, name)
				err := NetworkCreate(cli, realname, config)
				if err != nil {
					fmt.Println(err)
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

	}

	// # Exposed Ports
	// # Dependencies?
	// # More services config params 
	// # Timeouts / Errors

}

func ProjectName() string {
    // # Get stack name from --name
	// # Get stack name from directory if not passed 
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    _, dir := path.Split(pwd)

    r := regexp.MustCompile("[^a-z0-9]+")
    return r.ReplaceAllString(strings.ToLower(dir), "")
}

func NetworkCreate(cli client.APIClient, name string, network *config.NetworkConfig) error {
    fmt.Printf("Creating network %q with driver %q\n", name, network.Driver)
    _, err := cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
        CheckDuplicate: true,
        Driver: network.Driver,
    })

    return err
}
