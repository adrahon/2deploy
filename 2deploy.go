package main

import (
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
    "github.com/docker/docker/api/types/filters"
)

func main() {

    project_name := ProjectName()

    project := project.NewProject(&project.Context{
            ComposeFiles: []string{"docker-compose.yml"},
            ProjectName:  project_name,
    }, nil, &config.ParseOptions{})

    if err := project.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
    }

    cli, err := client.NewEnvClient()
    if err != nil {
		fmt.Println(err)
		os.Exit(1)
    }

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
                real_name := name
                if config.External.Name != "" {
                    real_name = config.External.Name
                }
                fmt.Printf("Checking if external network %q exists\n", real_name)
                err := CheckNetworkExists(cli, real_name)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

            } else {
                // else create network
                real_name := fmt.Sprintf("%s_%s", project_name, name)
				err := NetworkCreate(cli, real_name, config)
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
                fmt.Printf("Volume: %q (external)\n", name)
                // handle external name
                if config.External.Name != "" {
                    fmt.Printf("Volume: %q (external: %q)\n", name, config.External.Name)
                }
            } else {
                // # else create volume
                fmt.Printf("Volume: %q\n", name)
            }
        }
    }

	// # Services

    if project.ServiceConfigs == nil {
        // no services, abort
		fmt.Println("No services defined, aborting")
		os.Exit(1)
    } else {
        for name, config := range project.ServiceConfigs.All() {
            fmt.Printf("Service: %q\n", project.Name + "_" + name)
            if config.Image != "" {
                fmt.Printf("  Image: %q\n", config.Image)
            } else {
                // # if no image abort
                fmt.Println("  no image defined for service, aborting")
				os.Exit(1)
            }
            for _, port := range config.Ports {
                fmt.Printf("  Port: %q\n", port)
            }
            if config.Networks != nil && len(config.Networks.Networks) != 0 {
                for _, network := range config.Networks.Networks {
                    fmt.Printf("  Network: %q\n", network.RealName)
                }
            }
            if config.Volumes != nil && len(config.Volumes.Volumes) != 0 {
                for _, volume := range config.Volumes.Volumes {
                    fmt.Printf("  Volume: %q\n", volume)
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
    err := CheckNetworkExists(cli, name)
	if err != nil {
		_, err := cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
			CheckDuplicate: true,
			Driver: network.Driver,
		})
		return err
    } else {
        fmt.Printf("Network %q exists, skipping\n", name)
	}
    return err
}

func CheckNetworkExists(cli client.APIClient, name string) error {
    filter := filters.NewArgs()
    filter.Add("name", name)
	list_options := types.NetworkListOptions{
		Filters: filter,
	}
	networkResources, err := cli.NetworkList(context.Background(), list_options)
	if err != nil {
		return err
	}
	if len(networkResources) != 1 {
	    return fmt.Errorf("Network %s could not be found.", name)
	}
	return err
}
