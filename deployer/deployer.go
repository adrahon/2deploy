package deployer

import (
	"fmt"
    "golang.org/x/net/context"

	"github.com/docker/libcompose/config"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/api/types/swarm"
)

// Network holds information for one network
type Network struct {
    RealName string // name in swarm
    Config    config.NetworkConfig
}

// Service holds information for one service
type Service struct {
    RealName string // name in swarm
    Spec     swarm.ServiceSpec
}

// Deployer holds information for deploying the project
type Deployer struct {
    client   client.APIClient
    context  context.Context
    Networks map[string]Network
    Project  string
    Services map[string]Service
}

// NewDeployer creates a deployer
func NewDeployer(project string, client client.APIClient, context context.Context) *Deployer {
    d := &Deployer{
        client:  client,
        context: context,
        Networks: make(map[string]Network),
        Project: project,
        Services: make(map[string]Service),
    }

    return d
}

func (d *Deployer) NetworkCreate(name string) error {
    fmt.Printf("Creating network %q with driver %q\n", name, d.Networks[name].Config.Driver)
    err := d.CheckNetworkExists(name)
    if err != nil {
        _, err := d.client.NetworkCreate(d.context, name, types.NetworkCreate{
            CheckDuplicate: true,
            Driver: d.Networks[name].Config.Driver,
        })
        return err
    } else {
        fmt.Printf("Network %q exists, skipping\n", name)
    }
    return err
}

func (d *Deployer) CheckNetworkExists(name string) error {
    filter := filters.NewArgs()
    realname := d.Networks[name].RealName
    filter.Add("name", realname)
    list_options := types.NetworkListOptions{
        Filters: filter,
    }
    networkResources, err := d.client.NetworkList(d.context, list_options)
    if err != nil {
        return err
    }
    if len(networkResources) != 1 {
        return fmt.Errorf("Network %s could not be found.", name)
    }
    return err
}

func (d *Deployer) ServiceCreate(service string) (types.ServiceCreateResponse, error) {
    response, err := d.client.ServiceCreate(d.context, d.Services[service].Spec, types.ServiceCreateOptions{})
    return response, err
}


