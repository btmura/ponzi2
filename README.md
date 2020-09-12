# ![ponzi2 logo of a pyramid](internal/app/view/ui/data/icon.png) ponzi2

[ponzi2](https://ponzi2.io) is a stock chart viewer written in [Go](https://golang.org)
using [OpenGL](https://github.com/go-gl/gl) and [GLFW](https://github.com/go-gl/glfw/).

Visit [ponzi2.io](https://ponzi2.io) for videos and tutorials! Star the 
[repository](https://github.com/btmura/ponzi2)!

## Features

* View charts using data provided for free by [IEX](https://iextrading.com/developer).
  View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).
* Runs on both [Windows and Linux](https://github.com/btmura/ponzi2/releases).

## Getting Started

[![Build Status](https://travis-ci.org/btmura/ponzi2.svg?branch=master)](https://travis-ci.org/btmura/ponzi2)
[![Go Report Card](https://goreportcard.com/badge/github.com/btmura/ponzi2)](https://goreportcard.com/report/github.com/btmura/ponzi2)

Download a stable [release](https://github.com/btmura/ponzi2/releases) or run this command to install ponzi2:

`go get -u github.com/btmura/ponzi2/...`

If you have problems with Go, OpenGL, or GLFW, check out these OS-specific instructions: 

* For Windows, follow the [How to Setup Go, OpenGL, and GLFW on Windows](https://youtu.be/aeHfqk0cVOE) tutorial.
* For Fedora Core, check out the [Running ponzi2 on Linux FC27](https://github.com/btmura/ponzi2/issues/4) issue.

### Generating Code

The following tools are needed to generate code when making certain types of changes. 
All generated code is already checked-in, so you shouldn't have to do this most of the time.

1. `go get -u github.com/mjibson/esc` for packaging resources into the binary.
2. `go get -u github.com/akavel/rsrc` for converting the icon into a Windows resources file.
3. `go get -u golang.org/x/tools/cmd/stringer` for generating String() methods.

Execute `go generate ./...` to [generate](https://blog.golang.org/generate) the code.

## Getting Help

File an [issue](https://github.com/btmura/ponzi2/issues) for help.

## How to Contribute

* File [issues](https://github.com/btmura/ponzi2/issues) for feature requests and bugs.

* Send a pull request for the following: 

    * Improve the [Go Report Card](https://goreportcard.com/report/github.com/btmura/ponzi2) score.

## Credits

Thank you!

* btmura - Main developer
* ajd3v - Added trackline legend to price chart.

## Disclaimers

Data provided for free by provided for free by [IEX](https://iextrading.com/developer).
View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).
