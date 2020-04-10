package docker

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
	"github.com/jhoonb/archivex"
	"judgeBackend/src/service/sample"
	"os"
	"path"
	"sync/atomic"
	"time"
)

const (
	MaxContainerLimit = 12
	WaitDuration      = 5
)

var curNumContainer uint32 = 0
var building = mapset.NewSet()

func GetIPAddress(id string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	cli.NegotiateAPIVersion(context.Background())
	res, err := cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return "", err
	}
	return (*res.NetworkSettings).IPAddress, nil
}

func StartContainer(s sample.Sample, ports []nat.Port) (string, error) {

	for atomic.LoadUint32(&curNumContainer) >= MaxContainerLimit {
		time.Sleep(WaitDuration * time.Second)
	}
	atomic.AddUint32(&curNumContainer, 1)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	cli.NegotiateAPIVersion(context.Background())
	exposedPorts := nat.PortSet{}
	for _, p := range ports {
		exposedPorts[p] = struct{}{}
	}
	res, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image:        fmt.Sprintf("%s:latest", s.Tag()),
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{}, &network.NetworkingConfig{}, "")
	if err != nil {
		return "", err
	}
	err = cli.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	return res.ID, err
}

func RemoveContainer(id string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	err = cli.ContainerStop(context.Background(), id, nil)
	if err != nil {
		return err
	}
	atomic.AddUint32(&curNumContainer, ^uint32(0))
	return cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
}

func ImageExist(s sample.Sample) (bool, bool, error) {
	l, err := ImageList()
	if err != nil {
		return false, false, err
	}
	for _, subl := range l {
		for _, tag := range subl {
			if s.Tag()+":latest" == tag {
				return true, building.Contains(s.Tag()), nil
			}
		}
	}
	return false, building.Contains(s.Tag()), nil
}

func ImageList() ([][]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(context.Background())
	imageSummaryList, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	res := make([][]string, len(imageSummaryList))
	for i, s := range imageSummaryList {
		res[i] = s.RepoTags
	}
	return res, nil
}
func Build(s sample.Sample) error {
	building.Add(s.Tag())
	imageContextDir := path.Dir(s.Spec.DockerFile)
	tar := new(archivex.TarFile)
	err := tar.Create("/tmp/pg_context.tar")
	if err != nil {
		return err
	}
	err = tar.AddAll(imageContextDir, false)
	if err != nil {
		return err
	}
	err = tar.Close()
	if err != nil {
		return err
	}
	buildContext, err := os.Open("/tmp/pg_context.tar")
	defer buildContext.Close()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	args := map[string]*string{
		"DB_DUMP_FILE": &s.Spec.Database,
	}
	options := types.ImageBuildOptions{
		NoCache:        false,
		Tags:           []string{s.Tag()},
		PullParent:     true,
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		BuildArgs:      args,
		Dockerfile:     "Dockerfile",
	}
	buildResponse, err := cli.ImageBuild(context.Background(), buildContext, options)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	return jsonmessage.DisplayJSONMessagesStream(buildResponse.Body, os.Stderr, termFd, isTerm, nil)
}
