#version 450 core

layout(location = 10) uniform sampler2D texture;
layout(location = 11) uniform float colorMixAmount;

in vec2 texCoord;
in vec4 color;
in vec3 lighting;

out vec4 fragColor;

void main(void) {
	vec4 texColor = texture2D(texture, texCoord);
	vec4 mixColor = mix(texColor, color, colorMixAmount);
	fragColor = vec4(mixColor.rgb * lighting, mixColor.a);
}