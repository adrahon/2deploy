package main

import (
    "fmt"
    "os"
    "path"
    "regexp"
    "strings"
    "strconv"

    "golang.org/x/net/context"

    "github.com/adrahon/2deploy/deployer"

    "github.com/docker/libcompose/config"
    "github.com/docker/libcompose/project"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/swarm"
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

    deployer := deployer.NewDeployer(cli, context.Background())

    // # Check if stack exists

    // Networks

    default_network := ""
    if project.NetworkConfigs == nil || len(project.NetworkConfigs) == 0 {
        // if no network create default
        name := fmt.Sprintf("%s_default", project_name)
        config := config.NetworkConfig { Driver: "overlay", }
		err := deployer.NetworkCreate(name, &config)
		if err != nil {
			fmt.Println(err)
		}
        default_network = name
    } else {
        for name, config := range project.NetworkConfigs {
            // # if network external check if exists
            if config.External.External {
                real_name := name
                if config.External.Name != "" {
                    real_name = config.External.Name
                }
                fmt.Printf("Checking if external network %q exists\n", real_name)
                err := deployer.CheckNetworkExists(real_name)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

            } else {
                // else create network
                real_name := fmt.Sprintf("%s_%s", project_name, name)
				err := deployer.NetworkCreate(real_name, config)
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
			service_name := fmt.Sprintf("%s_%s", project_name, name)

            ports := []swarm.PortConfig{}
            for _, p := range config.Ports {
                port := strings.Split(p, ":") 
                if len(port) > 1 {
                    t, _ := strconv.Atoi(port[1])
                    p, _ := strconv.Atoi(port[0])
					ports = append(ports, swarm.PortConfig{
						TargetPort:    uint32(t),
						PublishedPort: uint32(p),
					})
                } else {
                    t, _ := strconv.Atoi(port[0])
					ports = append(ports, swarm.PortConfig{
						TargetPort:    uint32(t),
					})
                }
            }

			nets := []swarm.NetworkAttachmentConfig{}
            // use default network if exists
            if default_network != "" {
                nets = append(nets, swarm.NetworkAttachmentConfig{Target: default_network})
            } else {
                if config.Networks != nil && len(config.Networks.Networks) != 0 {
                    for _, network := range config.Networks.Networks {
                        nets = append(nets, swarm.NetworkAttachmentConfig{Target: network.RealName})
                    }
                }
            }

			service_spec := swarm.ServiceSpec{
				Annotations: swarm.Annotations{
					Name:   service_name,
				},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: swarm.ContainerSpec{
						Image:   config.Image,
						Command: config.Command,
						//Args:    service.Args,
						Env:     config.Environment,
					},
				},
				EndpointSpec: &swarm.EndpointSpec{
			    		Ports: ports,
				},
				Networks: nets,
			}

            fmt.Printf("Creating service %q\n", service_name)

            s, err := cli.ServiceCreate(context.Background(), service_spec, types.ServiceCreateOptions{})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
 
            fmt.Printf("ID: %s\n\n", s.ID)

            if config.Volumes != nil && len(config.Volumes.Volumes) != 0 {
                for _, volume := range config.Volumes.Volumes {
                    fmt.Printf("  Volume: %q\n", volume)
                }
            }
        }

	}

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

