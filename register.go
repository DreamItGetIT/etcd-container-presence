package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/dotcloud/docker/api"
	docker "github.com/dotcloud/docker/api/client"
)

const (
	dockerIP = "172.17.42.1"
	ttl      = 60
	timeout  = 30 * time.Second
)

type registerError struct {
	message string
}

func (e *registerError) Error() string {
	return e.message
}

type dockerInspectWriter struct {
	lastCommand []dockerInspectPortMapping
}

type dockerInspectPortMapping struct {
	HostConfig struct {
		PortBindings map[string][]struct {
			HostIp   string
			HostPort string
		}
	}
	State struct {
		Running bool
	}
}

type dockerPortMapping struct {
	ContainerPort string
	Port          string
	Host          string
}

func (dpr dockerInspectPortMapping) portMappingsList() []*dockerPortMapping {
	dockerPortMappings := make([]*dockerPortMapping, 0, len(dpr.HostConfig.PortBindings))

	for ContainerPort, Binding := range dpr.HostConfig.PortBindings {
		currentDockerPortMapping := dockerPortMapping{}

		if len(Binding) == 0 {
			continue
		}

		if pos := strings.Index(ContainerPort, "/"); pos >= 0 {
			currentDockerPortMapping.ContainerPort = ContainerPort[:pos]
		} else {
			currentDockerPortMapping.ContainerPort = ContainerPort
		}

		if pos := strings.Index(Binding[0].HostPort, "/"); pos >= 0 {
			currentDockerPortMapping.Port = Binding[0].HostPort[:pos]
		} else {
			currentDockerPortMapping.Port = Binding[0].HostPort
		}

		if Binding[0].HostIp == "0.0.0.0" {
			currentDockerPortMapping.Host = dockerIP
		} else {
			currentDockerPortMapping.Host = Binding[0].HostIp
		}

		dockerPortMappings = append(dockerPortMappings, &currentDockerPortMapping)
	}

	return dockerPortMappings
}

func (diw *dockerInspectWriter) Write(p []byte) (n int, err error) {
	json.Unmarshal(p, &diw.lastCommand)
	return len(p), nil
}

func startRegistration(c *cli.Context) {
	if !c.IsSet("container") {
		fmt.Println("--container argument is required")
		return
	}

	go deregister(c.GlobalString("container"))
	fmt.Println("registereing container")

	for {
		if err := register(c.GlobalString("container")); err != nil {
			fmt.Fprintf(os.Stderr, "registration failed ", err)
		}

		time.Sleep(timeout)
	}
}

func getContainerInfo(container string) (*dockerInspectPortMapping, error) {
	dockerWriter := dockerInspectWriter{}
	dockerClient := docker.NewDockerCli(nil, &dockerWriter, os.Stderr, "unix", api.DEFAULTUNIXSOCKET, nil)
	dockerClient.CmdInspect(container)

	if len(dockerWriter.lastCommand) == 0 {
		return nil, &registerError{message: "Container does not exist"}
	}

	if dockerWriter.lastCommand[0].State.Running == false {
		return nil, &registerError{message: "Container is not running"}
	}

	return &dockerWriter.lastCommand[0], nil
}

func register(container string) error {
	etcdClient := etcd.NewClient([]string{fmt.Sprintf("http://%s:4001", dockerIP)})
	containerInfo, err := getContainerInfo(container)

	if err != nil {
		return err
	}

	if _, err := etcdClient.UpdateDir(fmt.Sprint("containers/", container), ttl); err != nil {
		// If update dir fails is because the directory doesn't exist, so, let's create it
		if _, err := etcdClient.SetDir(fmt.Sprint("containers/", container), ttl); err != nil {
			return err
		}
	}

	for _, dockerPortMapping := range containerInfo.portMappingsList() {
		if _, err := etcdClient.Set(fmt.Sprintf("containers/%s/ports/%s/host/", container, dockerPortMapping.ContainerPort), dockerPortMapping.Host, ttl); err != nil {
			return err
		}

		if _, err := etcdClient.Set(fmt.Sprintf("containers/%s/ports/%s/port/", container, dockerPortMapping.ContainerPort), dockerPortMapping.Port, ttl); err != nil {
			return err
		}
	}

	return nil
}

func deregister(container string) error {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch

	etcdClient := etcd.NewClient([]string{fmt.Sprintf("http://%s:4001", dockerIP)})

	if _, err := etcdClient.Delete(fmt.Sprint("containers/", container), true); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func init() {

}

func main() {
	app := cli.NewApp()
	app.Name = "register"
	app.Usage = "Register the ports of a specfied Docker container with Etcd"
	app.Action = startRegistration
	app.Version = "0.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "container", Usage: "The container name or id"},
	}

	app.Run(os.Args)
}
