# Modbridge

[![Build Status](https://travis-ci.com/mhemeryck/modbridge.svg?branch=master)](https://travis-ci.com/mhemeryck/modbridge)

A tool for polling a modbus instance and pushing updates over MQTT.

## Quickstart

The recommend way is to build everything using [docker] and [docker-compose].

Building:

    docker-compose build

Running the application:

    docker-compose up modbridge

Note that the current setup still has all settings hard-coded, see the `main.go` file for the settings to change.

### Docker

The current docker setup uses a [multi-stage build], where the first stage build all the artifacts and still has all the test dependencies available, where as the second stage only contains the raw binary to execute (low footprint image).

## Testing

Testing is limited at the moment, but can be executed within the docker container:

    docker-compose run --rm modbridge_build go test -v ./...


[docker]: https://docs.docker.com/get-started/
[docker-compose]: https://docs.docker.com/compose/gettingstarted/#prerequisites
[multi-stage build]: https://docs.docker.com/develop/develop-images/multistage-build/
