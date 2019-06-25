**This repo is based on [EdgeXFoundry Export Services](https://github.com/edgexfoundry/edgex-go/tree/delhi)**

# What's new?

To allow Edgexfoundry export data to MindSphere(Siemens open IoT operating systhem), some codes of export-client and export-distro services need to be rewritten. Two feasible ways at present, one is by Cumulocity, another is using Open Edge Device Kit(OEDK).

## Cumulocity

1. Add `MINDCONNECT_TOPIC` as the topic of Cumulocity and rewrite related detection code.
2. Add `internal/export/distro/mindconnect.go` file to exec specific export commands.

## Open Edge Device Kit

1. [OEDK Documentation](https://developer.mindsphere.io/resources/openedge-devicekit/index.html) here.
2. Add `OEDKCONNECT_TOPIC` as the topic of OEDK and rewrite related detection code.
3. Add `cmd/export-distro/res/oedkconfig.toml` for some pre-configurations
4. Add `internal/export/distro/loadandupdateconfig.go` file for oedk configuration, `internal/export/distro/oedktopic.go` file for oedk topics, `internal/export/distro/oedkconnect.go` file to exec specific export commands.

## Export data to OEDK
1. Install and run OEDK and Mosquitto, refer to documentation.
2. Run the export-client service and update a new export endpoint by RESTful API of export-client service, use mosquitto address and port, and use `OEDKCONNECT_TOPIC`.
3. Do this step iff OEDK hasn't been initialized. Download onboard key and copy the context to `oedkconfig.toml` in single quotes, that is due to the escape character. Change the IsInitialized option to false. Ignore the DataSorceId and DataPointId.
4. Run the export-distro.

**Note:** Only export-distro doesn't support cross compiler. If needed, please built on arm64.

----
The following is the offical README

# EdgeX Foundry Services
[![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/export-go)](https://goreportcard.com/report/github.com/edgexfoundry/export-go)
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

EdgeX Foundry is a vendor-neutral open source project hosted by The Linux Foundation building a common open framework for IoT edge computing.  At the heart of the project is an interoperability framework hosted within a full hardware- and OS-agnostic reference software platform to enable an ecosystem of plug-and-play components that unifies the marketplace and accelerates the deployment of IoT solutions.  This repository contains the Go implementation of EdgeX Foundry microservices.  It also includes files for building the services, containerizing the services, and initializing (bootstrapping) the services.

# Install and Deploy Native

## Prerequisites
### pkg-config
`go get github.com/rjeczalik/pkgconfig/cmd/pkg-config`

### ZeroMQ
Several EdgeX Foundry services depend on ZeroMQ for communications by default.

The easiest way to get and install ZeroMQ on Linux is to use or follow the following setup script:  https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00.

For MacOS, use brew: `brew install zeromq`. Please note that the necessary `pc` file will need to be added to the `PKG_CONFIG_PATH` environment variable. For example `PKG_CONFIG_PATH=/usr/local/Cellar/zeromq/4.2.5/lib/pkgconfig/`

**Note**: Setup of the ZeroMQ library is not supported on Windows plaforms.

## Installation and Execution
To fetch the code and build the microservice execute the following:

```
cd $GOPATH/src
go get github.com/edgexfoundry/export-go
cd $GOPATH/src/github.com/edgexfoundry/export-go
# pull the 3rd party / vendor packages
make prepare
# build the microservices
make build
# run the services
make run
```

**Note** You will need to have the database running before you execute `make run`. If you don't want to install a database locally, you can bring one up via their respective Docker containers.

# Install and Deploy via Docker Container #
This project has facilities to create and run Docker containers.

### Prerequisites ###
See https://docs.docker.com/install/ to learn how to obtain and install Docker.

### Installation and Execution ###

```
cd $GOPATH/src
go get github.com/edgexfoundry/export-go
cd $GOPATH/src/github.com/edgexfoundry/export-go
# To create the Docker images
sudo make docker
# To run the containers
sudo make run_docker
```

# Install and Deploy via Snap Package #
EdgeX Foundry is also available as a snap package, for more details
on the snap, including how to install it, please refer to [EdgeX snap](https://github.com/edgexfoundry/export-go/blob/master/snap/README.md)

# Docker Hub #
EdgeX images are kept on organization's [DockerHub page](https://hub.docker.com/u/edgexfoundry/).
They can be run in orchestration via official [docker-compose.yml](https://github.com/edgexfoundry/developer-scripts/blob/master/compose-files/docker-compose.yml).

The simplest way is to do this via prepared script in `bin` directory:
```
cd bin 
./edgex-docker-launch.sh
```

### Compiled Binaries
During development phase, it is important to run compiled binaries (not containers).

There is a script in `bin` directory that can help you launch the whole EdgeX system:
```
cd bin
./edgex-launch.sh
```

## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)
