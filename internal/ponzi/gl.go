package ponzi

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

func glCreateProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vs, err := glCreateShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fs, err := glCreateShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
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

		return 0, fmt.Errorf("glCreateProgram: failed to create program: %q", log)
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return p, nil
}

func glCreateShader(shaderSource string, shaderType uint32) (uint32, error) {
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

		return 0, fmt.Errorf("glCreateShader: failed to compile shader, type: %d, source: %q, log: %q", shaderType, src, logLen)
	}

	return sh, nil
}
