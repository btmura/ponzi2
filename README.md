# ponzi2

[![Build Status](https://travis-ci.org/btmura/ponzi2.svg?branch=master)](https://travis-ci.org/btmura/ponzi2)

[ponzi2](https://ponzi2.io) is a stock chart viewer written in [Go](https://golang.org) using [OpenGL](https://github.com/go-gl/gl) and [GLFW](https://github.com/go-gl/glfw/).

It is go-gettable:

`go get -u github.com/btmura/ponzi2`

## Stock Data

Data provided for free by [IEX](https://iextrading.com/developer). View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).

## Development Environment Setup

* `go get -u github.com/mjibson/esc`
* `go get -u github.com/akavel/rsrc`
* `go get -u golang.org/x/tools/cmd/stringer`
* Setup `protoc` compiler.
* `go generate ./...`
