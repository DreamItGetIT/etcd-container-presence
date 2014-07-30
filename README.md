# etcd presence container

An opinionated [Docker](https://www.docker.com/) image which contains a simple program which checks if the specified container is running and register its exposed ports to a running [etcd](http://coreos.com/using-coreos/etcd/) local instance.

It is a [Go](http://golang.org/) implementation of the original [Python one](https://github.com/mwagg/etcd-container-presence), written by [Mike Wagg](https://github.com/mwagg).

## Description

An opinionated [Docker](https://www.docker.com/) image which contains a simple program which checks if a container is running and in that case it registers the mapped ports to a __local [etcd](http://coreos.com/using-coreos/etcd/) instance__.

The __local [etcd](http://coreos.com/using-coreos/etcd/) instance__ must run into a docker container, which must expose the [etcd](http://coreos.com/using-coreos/etcd/) client default port 4001 in all the network interfaces, IP `0.0.0.0`, therefore it is accessible under the IP `172.17.42.1` (default docker daemon IP).

The requirement of [etcd](http://coreos.com/using-coreos/etcd/) is because the `register` program included in this repository, whose source code is in the root, gets the specified [Docker](https://www.docker.com/) container, whose name is passed as a command line argument, checks if it is running and in that case, registers the exposed ports into [etcd](http://coreos.com/using-coreos/etcd/) instance the following keys with a 60 seconds of `ttl` (time to live):

* Create the directory `/containers/{container_name}/` if it doesn't exist, where `{container_name}` is the specified container's name
* For each port register the following keys into the created directory:
    * `/containers/{container_name}/{port}/host` and whose value is the IP of the interface where the port is exposed with a fallback to the default docker daemon IP (`172.17.42.1`) when it is `0.0.0.0`
    * `/containers/{container_name}/{port}/port` and whose value is the port exposed by docker to the host machine
     
    NOTE: that `{port}` is the port exposed by the container.

`register` muse be run with the command line parameter `--container {container_anme}` where `{container_name}` is the container's name to monitor; it will check every 30 seconds the specified container and if it is still running, updates the keys so the keys get again the 60 seconds of `ttl`, otherwise it doesn't do anything and the keys will be deleted automatically by [etcd](http://coreos.com/using-coreos/etcd/) when the `ttl` expires.

When `register` is stopped by sending `SIGTERM` signal, it unregisters the container, removing the mentioned keys and directory from [etcd](http://coreos.com/using-coreos/etcd/).
    

## How to use

Run a container which monitors the named container `my_container`

`docker run -rm -v /run/docker.sock:/run/docker.sock digit/etcd-container-presence --container my_container`

We use the following [docker's run parameters](https://docs.docker.com/reference/commandline/cli/#run)

`--rm`  to remove the container when it stops. No reason for it to stick around
`-v`    to map the [Docker](https://www.docker.com/) socket into the container so it can query the Docker API

When the container be stopped, the directory and keys registered in [etcd](http://coreos.com/using-coreos/etcd/) as `register` program does.

## License

Just MIT, Copyright (c) 2014 DreamItGetIT, read LICENSE file for more information.

