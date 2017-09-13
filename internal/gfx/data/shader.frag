#version 450 core

layout(location = 5) uniform sampler2D texture;
layout(location = 6) uniform float colorMixAmount;
layout(location = 7) uniform vec3 textColor;

in vec2 texCoord;
in vec4 color;

out vec4 fragColor;

void main(void) {
	vec4 texColor = vec4(textColor, 1.0) * vec4(1.0, 1.0, 1.0, texture2D(texture, texCoord).r);
	fragColor = mix(texColor, color, colorMixAmount);
}