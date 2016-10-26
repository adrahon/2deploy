package deployer

import (
	"fmt"
    "golang.org/x/net/context"

	"github.com/docker/libcompose/config"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// Deployer holds information for deploying the project
type Deployer struct {
    client  client.APIClient
    context context.Context
}

// NewDeployer creates a deployer
func NewDeployer(client client.APIClient, context context.Context) *Deployer {
    d := &Deployer{
        client:  client,
        context: context,
    }

    return d
}

func (d *Deployer) NetworkCreate(name string, network *config.NetworkConfig) error {
    fmt.Printf("Creating network %q with driver %q\n", name, network.Driver)
    err := d.CheckNetworkExists(name)
    if err != nil {
        _, err := d.client.NetworkCreate(d.context, name, types.NetworkCreate{
            CheckDuplicate: true,
            Driver: network.Driver,
        })
        return err
    } else {
        fmt.Printf("Network %q exists, skipping\n", name)
    }
    return err
}

func (d *Deployer) CheckNetworkExists(name string) error {
    filter := filters.NewArgs()
    filter.Add("name", name)
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

