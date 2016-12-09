#version 440 core

layout(location = 0) uniform mat4 projectionViewMatrix;
layout(location = 1) uniform mat4 modelMatrix;
layout(location = 2) uniform mat4 normalMatrix;

layout(location = 3) uniform vec3 ambientLightColor;
layout(location = 4) uniform vec3 directionalLightColor;
layout(location = 5) uniform vec3 directionalVector;

layout(location = 6) in vec4 position;
layout(location = 7) in vec4 normal;
layout(location = 8) in vec2 inTexCoord;

out vec2 texCoord;
out vec3 lighting;

void main(void) {
	gl_Position = projectionViewMatrix * modelMatrix * position;

	texCoord = inTexCoord;

	vec4 transformedNormal = normalMatrix * vec4(normal.xyz, 1.0);
	float directional = max(dot(transformedNormal.xyz, directionalVector), 0.0);
	lighting = ambientLightColor + (directionalLightColor * directional);
}