#version 450 core

layout(location = 0) uniform mat4 projectionViewMatrix;
layout(location = 1) uniform mat4 modelMatrix;

layout(location = 2) in vec4 position;
layout(location = 3) in vec4 inColor;
layout(location = 4) in vec2 inTexCoord;

out vec4 color;
out vec2 texCoord;

void main(void) {
	gl_Position = projectionViewMatrix * modelMatrix * position;
	color = inColor;
	texCoord = inTexCoord;
}