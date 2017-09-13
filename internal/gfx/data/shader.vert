#version 450 core

layout(location = 0) uniform mat4 projectionViewMatrix;
layout(location = 1) uniform mat4 modelMatrix;

layout(location = 2) in vec4 position;
layout(location = 3) in vec2 inTexCoord;
layout(location = 4) in vec4 inColor;

out vec2 texCoord;
out vec4 color;

void main(void) {
	gl_Position = projectionViewMatrix * modelMatrix * position;
	texCoord = inTexCoord;
	color = inColor;
}