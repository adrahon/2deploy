package main

import (
    "flag"
    "fmt"
    "os"
    "path"
    "regexp"
    "strings"
    "strconv"

    "golang.org/x/net/context"

    "github.com/adrahon/deploy-compose/deployer"

    "github.com/docker/libcompose/config"
    "github.com/docker/libcompose/project"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types/swarm"
    mounttypes "github.com/docker/docker/api/types/mount"
)

var projectFlag = flag.String("p", "", "Specify an alternate project name (default: directory name)")
var fileFlag = flag.String("f", "docker-compose.yml", "Specify an alternate compose file")

func main() {

    // Process command and parameters

    flag.Usage = usage

    flag.Parse()
    project_name := *projectFlag
    if project_name == "" {
        project_name = ProjectName()
    }

    compose_file := *fileFlag

    project := project.NewProject(&project.Context{
        ComposeFiles: []string{compose_file},
        ProjectName:  project_name,
    }, nil, &config.ParseOptions{})

    command := "usage"
    if len(flag.Args()) > 0 {
        command = flag.Args()[0]
    }

    // Load compose file
    if err := project.Parse(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // Initialize Docker client
    cli, err := client.NewEnvClient()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    deployer := deployer.NewDeployer(project_name, cli, context.Background())
    initDeployer(deployer, project)

    // Select command to run
    switch command {
    case "config":
        fmt.Println("command: ", command)
    case "create":
        fmt.Println("command: ", command)
    case "down":
        down(deployer)
    case "help":
        usage()
    case "restart":
        fmt.Println("command: ", command)
    case "rm":
        fmt.Println("command: ", command)
    case "up":
        up(deployer)
    default:
        fmt.Fprintf(os.Stderr, "No such command: %s\n", command)
        usage()
    }
}

func initDeployer(d *deployer.Deployer, project *project.Project) {

    // Networks
    if project.NetworkConfigs == nil || len(project.NetworkConfigs) == 0 {
        // if no network create default
        name := fmt.Sprintf("%s_default", d.Project)
        config := config.NetworkConfig { Driver: "overlay", }
        network := deployer.Network { RealName: name, Config: config }
        d.Networks["default"] = network
    } else {
        for name, config := range project.NetworkConfigs {
            realname := name
            // if network external keep name
            if config.External.External {
                if config.External.Name != "" {
                    realname = config.External.Name
                }
            } else {
                // namespace name
                realname = fmt.Sprintf("%s_%s", d.Project, name)
            }
            network := deployer.Network { RealName: realname, Config: *config}
            d.Networks[name] = network
        }
    }

    // # Volumes
    if project.VolumeConfigs != nil && len(project.VolumeConfigs) != 0 {
        for name, config := range project.VolumeConfigs {
            // # if volume external check if exists
            if config.External.External {
                fmt.Printf("Volume: %q (external)\n", name)
                // handle external name
                if config.External.Name != "" {
                    fmt.Printf("Volume: %q (external: %q)\n", name, config.External.Name)
                }
            } else if config.Driver != "" {
                // # else create volume ?
                fmt.Printf("Volume: %q\n", name)
            }
        }
    }

    // # Services
    if project.ServiceConfigs != nil {
        for name, config := range project.ServiceConfigs.All() {
            realname := fmt.Sprintf("%s_%s", d.Project, name)

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
            if config.Networks == nil || len(config.Networks.Networks) == 0 {
                // This never happens?
                // if default network defined use for service
                network := d.Networks[fmt.Sprintf("%s_default", d.Project)] // 🤔
                if network.RealName != "" {
                    nets = append(nets, swarm.NetworkAttachmentConfig{Target: network.RealName})
                }
            } else {
                for _, network := range config.Networks.Networks {
                    nets = append(nets, swarm.NetworkAttachmentConfig{Target: d.Networks[network.Name].RealName})
                }
            }

            mounts := []mounttypes.Mount{}
            if config.Volumes != nil && len(config.Volumes.Volumes) != 0 {
                for _, volume := range config.Volumes.Volumes {
                    mounts = append(mounts, mounttypes.Mount{ Type: mounttypes.TypeVolume, Target: volume.Destination, })
                }
            }

            service_spec := swarm.ServiceSpec{
                Annotations: swarm.Annotations{
                    Name:   realname,
                },
                TaskTemplate: swarm.TaskSpec{
                    ContainerSpec: swarm.ContainerSpec{
                        Image:   config.Image,
                        Command: config.Command,
                        // Args:    service.Args,
                        Env:     config.Environment,
                        // Labels:  runconfigopts.ConvertKVStringsToMap(opts.containerLabels.GetAll()),
                        // Dir:             opts.workdir,
                        // User:            opts.user,
                        // Groups:          opts.groups,
                        Mounts:  mounts,
                        // StopGracePeriod: opts.stopGrace.Value(),
                    },
                    // Networks:      convertNetworks(opts.networks),
                    // Resources:     opts.resources.ToResourceRequirements(),
                    // RestartPolicy: opts.restartPolicy.ToRestartPolicy(),
                    // Placement: &swarm.Placement{
                    //     Constraints: opts.constraints,
                    //},
                    // LogDriver: opts.logDriver.toLogDriver(),
                },
                EndpointSpec: &swarm.EndpointSpec{
                    Ports: ports,
                },
                Networks: nets,
            }

            service := deployer.Service { RealName: realname, Spec: service_spec}
            d.Services[name] = service

        }
    }
}

func down(deployer *deployer.Deployer) {

    // Services
    for name, _ := range deployer.Services {
        fmt.Printf("Removing service %q\n", deployer.Services[name].RealName)
        err := deployer.ServiceRemove(name)
        if err != nil {
            fmt.Println(err)
        }
    }

    // Networks
    // TODO check there are no remaining endpoints
    for name, network := range deployer.Networks {
        // remove if not external
        if !network.Config.External.External {
            fmt.Printf("Removing  network %q\n", deployer.Networks[name].RealName)
            err := deployer.NetworkRemove(name)
            if err != nil {
                fmt.Println(err)
            }
        }
    }



}

func usage() {
    fmt.Printf("A utiliy to deploy services defined in a compose file to swarm-mode clusters.\n")
    fmt.Printf("\nUsage:\n")
    fmt.Printf("  %s [options] [COMMAND]\n", os.Args[0])
    fmt.Printf("  %s -h|--help\n", os.Args[0])
    fmt.Printf("\nOptions:\n")
    flag.PrintDefaults()
    fmt.Printf("\nCommands:\n")
    fmt.Printf("  up                 Create and start services\n")
}

func up(deployer *deployer.Deployer) {

    // TODO Check if stack exists

    // Networks
    for name, network := range deployer.Networks {
        // if network external check if exists
        if network.Config.External.External {
            fmt.Printf("Checking if external network %q exists\n", name)
            err := deployer.CheckNetworkExists(name)
            if err != nil {
                fmt.Println(err)
                os.Exit(1)
            }
        } else {
            // else create network
            fmt.Printf("Creating network %q with driver %q\n", deployer.Networks[name].RealName, deployer.Networks[name].Config.Driver)
            err := deployer.NetworkCreate(name)
            if err != nil {
                fmt.Println(err)
            }
        }
    }

    // Services
    for name, _ := range deployer.Services {
        fmt.Printf("Creating service %q\n", deployer.Services[name].RealName)
        _, err := deployer.ServiceCreate(name)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
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

