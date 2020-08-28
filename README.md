![](phono.png)

[![GoDoc](https://godoc.org/pipelined.dev/phono?status.svg)](https://godoc.org/pipelined.dev/phono)
[![Build Status](https://travis-ci.org/pipelined/phono.svg?branch=master)](https://travis-ci.org/pipelined/phono)
[![Go Report Card](https://goreportcard.com/badge/pipelined.dev/phono)](https://goreportcard.com/report/pipelined.dev/phono)
[![codecov](https://codecov.io/gh/pipelined/phono/branch/master/graph/badge.svg)](https://codecov.io/gh/pipelined/phono)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpipelined%2Fphono.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpipelined%2Fphono?ref=badge_shield)

`phono` is a command for audio processing. It's build on top of [pipelined DSP framework](https://github.com/pipelined/pipe).

## Installation

Prerequisites:

* [lame](http://lame.sourceforge.net/) to enable mp3 encoding

To link lame from custom location, set `CGO_CFLAGS=-I<path-to-lame.h>` environment variable.

`go get pipelined.dev/phono`

## Usage

`phono encode` allows to decode/encode various audio files in cli or interactive web UI mode.

## Contributing

For a complete guide to contributing to `phono`, see the [Contribution guide](https://pipelined.dev/phono/blob/master/CONTRIBUTING.md).

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpipelined%2Fphono.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpipelined%2Fphono?ref=badge_large)