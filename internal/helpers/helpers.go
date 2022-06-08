package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

func StopDB(id string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	err = cli.ContainerKill(context.Background(), id, "")
	if err != nil {
		return err
	}

	return nil
}

func StartDB() (dburl string, id string, err error) {
	image := "postgres:latest"
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", "", err
	}

	imgFilter := filters.NewArgs()
	imgFilter.Add("reference", image)
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
	if err != nil {
		return "", "", err
	}
	if len(images) == 0 {
		_, err = cli.ImagePull(context.Background(), "docker.io/library/"+image, types.ImagePullOptions{})
		if err != nil {
			return "", "", err
		}
		var count = 0
		fmt.Print("Pulling")
		for count == 0 {
			images, err = cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
			if err != nil {
				return "", "", err
			}
			count = len(images)
			time.Sleep(time.Second)
			fmt.Print("*")
		}
		fmt.Println("Done")

	}
	containerConfig := &container.Config{
		Image: image,
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=user",
		},
	}
	//delay for pull
	time.Sleep(time.Second * 3)
	hostConfig := &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
	}

	resp, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", "", err
	}
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", err
	}
	filter := filters.NewArgs()
	filter.Add("id", resp.ID)
	var C types.Container
	for {
		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: filter})
		if err != nil {
			return "", "", err
		}
		if strings.Contains(containers[0].Status, "Up 1") {
			C = containers[0]
			time.Sleep(time.Second)
			break
		}
	}
	return fmt.Sprintf("postgres://user:password@127.0.0.1:%v", C.Ports[0].PublicPort), resp.ID, nil
}

// DBConnect .
func DBConnect() (dockerID string, client *pgxpool.Pool, cancel context.CancelFunc, err error) {
	var dburl string
	dburl, dockerID, err = StartDB()
	if err != nil {
		return "", nil, nil, err
	}
	client, cancel, err = Connect(dburl)
	if err != nil {
		return "", nil, nil, err
	}
	return dockerID, client, cancel, nil
}

// DBClose .
func DBClose(id string, cancel context.CancelFunc) error {
	cancel()
	return StopDB(id)
}

// Connect .
func Connect(dburl string) (*pgxpool.Pool, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client, err := pgxpool.Connect(ctx, dburl)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	logrus.Info("Database is connected")
	return client, cancel, nil
}
