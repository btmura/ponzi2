# ponzi2

ponzi2 is a stock chart viewer written in Go using OpenGL and GLFW. 

It is go-gettable:

`go get -u github.com/btmura/ponzi2`

## Stock Data

Data provided for free by [IEX](https://iextrading.com/developer). View [IEX’s Terms of Use](https://iextrading.com/api-exhibit-a/).

## Development Environment Setup

* `go get -u github.com/mjibson/esc`
* `go get -u golang.org/x/tools/cmd/stringer`
* Setup `protoc` compiler.
* `go generate ./...`
