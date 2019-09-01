# ![ponzi2 logo of a pyramid](internal/app/view/ui/data/icon.png) ponzi2

[ponzi2](https://ponzi2.io) is a stock chart viewer written in [Go](https://golang.org) using [OpenGL](https://github.com/go-gl/gl) and [GLFW](https://github.com/go-gl/glfw/).

Visit [ponzi2.io](https://ponzi2.io) for more videos and tutorials!

## Features

* View charts using data provided for free by [IEX](https://iextrading.com/developer). View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).
* Learn [Go](https://golang.org) from the actual code with [detailed tutorials](https://ponzi2.io/tutorials/)!
* Runs on both [Windows and Linux](https://github.com/btmura/ponzi2/releases).

## Getting Started

Run this command to install "ponzi2" and "ponzi2server".

`go get -u github.com/btmura/ponzi2/...`

### Development Environment Setup

[![Build Status](https://travis-ci.org/btmura/ponzi2.svg?branch=master)](https://travis-ci.org/btmura/ponzi2)

1. `go get -u github.com/mjibson/esc`
2. `go get -u github.com/akavel/rsrc`
3. `go get -u golang.org/x/tools/cmd/stringer`
4. `go generate ./...`

## Getting Help

Send an e-mail to btmura (address on GitHub profile).

## How to Contribute

Send a pull request.

## Credits

Thank you!

* btmura - Main developer
* ajd3v - Added trackline legend to price chart.

## Disclaimers

Data provided for free by provided for free by [IEX](https://iextrading.com/developer). View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).
