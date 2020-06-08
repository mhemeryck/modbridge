# Modbridge

[![Build Status](https://travis-ci.com/mhemeryck/modbridge.svg?branch=master)](https://travis-ci.com/mhemeryck/modbridge)
[![Coverage Status](https://coveralls.io/repos/github/mhemeryck/modbridge/badge.svg?branch=master)](https://coveralls.io/github/mhemeryck/modbridge?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mhemeryck/modbridge)](https://goreportcard.com/report/github.com/mhemeryck/modbridge)

A tool for polling a modbus instance and pushing updates over MQTT.

## Quickstart

Setup uses plain [golang build] for building the binaries.
Pre-built binaries are availble on the [releases] page.
Docker images shall be made available on [docker hub].

Custom configuration can be added by changing the `config.yml` file.
The current example configuration is for a [unipi neuron L303].
Currently only polling the coil values is implemented.


[golang build]: https://golang.org/pkg/go/build/
[releases]: https://github.com/mhemeryck/modbridge/releases/
[docker hub]: https://hub.docker.com/r/mhemeryck/modbridge/
[unipi neuron L303]: https://www.unipi.technology/unipi-neuron-l303-p23/
