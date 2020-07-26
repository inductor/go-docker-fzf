package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/ktr0731/go-fuzzyfinder"
)

type Container struct {
	ID   string
	Name string
}

var cli *client.Client

func containers() ([]Container, error) {
	args := filters.NewArgs()
	args.Add("status", "exited")
	args.Add("status", "paused")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	var containerList []Container

	for _, container := range containers {
		containerList = append(containerList, Container{ID: container.ID[:10], Name: container.Names[0][1:]})
	}

	return containerList, nil
}

func main() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}
	cs, err := containers()
	if err != nil {
		log.Fatal(err)
	}
	idx, err := fuzzyfinder.FindMulti(cs, func(i int) string {
		return fmt.Sprintf("%s %s", cs[i].ID, cs[i].Name)
	}, fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
		cJson, err := cli.ContainerInspect(context.Background(), cs[i].ID)
		if err != nil {
			return err.Error()
		}
		return cJson.Created
	}))
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range idx {
		err := cli.ContainerStart(context.Background(), cs[i].Name, types.ContainerStartOptions{})
		if err != nil {
			log.Println(err)
		} else {
			log.Println(cs[i].Name)
		}
	}
}
