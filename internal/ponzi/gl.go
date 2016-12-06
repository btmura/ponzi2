package ponzi

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

func createProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vs, err := createShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fs, err := createShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	p := gl.CreateProgram()
	gl.AttachShader(p, vs)
	gl.AttachShader(p, fs)
	gl.LinkProgram(p)

	var status int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(p, gl.INFO_LOG_LENGTH, &logLen)

		log := strings.Repeat("\x00", int(logLen)+1)
		gl.GetProgramInfoLog(p, logLen, nil, gl.Str(log))

		return 0, fmt.Errorf("createProgram: failed to create program: %q", log)
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return p, nil
}

func createShader(shaderSource string, shaderType uint32) (uint32, error) {
	sh := gl.CreateShader(shaderType)
	src, free := gl.Strs(shaderSource + "\x00")
	defer free()
	gl.ShaderSource(sh, 1, src, nil)
	gl.CompileShader(sh)

	var status int32
	gl.GetShaderiv(sh, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(sh, gl.INFO_LOG_LENGTH, &logLen)

		log := strings.Repeat("\x00", int(logLen)+1)
		gl.GetShaderInfoLog(sh, logLen, nil, gl.Str(log))

		return 0, fmt.Errorf("createShader: failed to compile shader, type: %d, source: %q, log: %q", shaderType, src, log)
	}

	return sh, nil
}

func createArrayBuffer(data []float32) uint32 {
	var name uint32
	gl.GenBuffers(1, &name)
	gl.BindBuffer(gl.ARRAY_BUFFER, name)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4 /* total bytes */, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	return name
}

func createElementArrayBuffer(data []uint16) uint32 {
	var name uint32
	gl.GenBuffers(1, &name)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, name)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(data)*2 /* total bytes */, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	return name
}
