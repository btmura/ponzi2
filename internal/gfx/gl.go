package gfx

import (
	"fmt"
	"image"
	_ "image/png" // Needed to decode PNG images.
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// program creates a program from vertex and fragment shader source code.
func program(vertexShaderSrc, fragmentShaderSrc string) (uint32, error) {
	vs, err := shader(vertexShaderSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fs, err := shader(fragmentShaderSrc, gl.FRAGMENT_SHADER)
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

func shader(shaderSource string, shaderType uint32) (uint32, error) {
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

		return 0, fmt.Errorf("createShader: failed to compile shader:\n\ntype: %d\n\nsource: %q\n\nlog: %q", shaderType, shaderSource, log)
	}

	return sh, nil
}

// texture creates a texture from an image.
func texture(rgba *image.RGBA) uint32 {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return tex
}

// arrayBuffer creates an array buffer from a slice of floats.
func arrayBuffer(data []float32) uint32 {
	var name uint32
	gl.GenBuffers(1, &name)
	gl.BindBuffer(gl.ARRAY_BUFFER, name)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4 /* total bytes */, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	return name
}

// elementArrayBuffer creates an element array buffer from a slice of floats.
func elementArrayBuffer(data []uint16) uint32 {
	var name uint32
	gl.GenBuffers(1, &name)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, name)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(data)*2 /* total bytes */, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	return name
}
