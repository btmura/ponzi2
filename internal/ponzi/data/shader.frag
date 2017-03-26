#version 450 core

layout(location = 10) uniform sampler2D texture;
layout(location = 11) uniform float colorMixAmount;
layout(location = 12) uniform vec3 textColor;

in vec2 texCoord;
in vec4 color;
in vec3 lighting;

out vec4 fragColor;

void main(void) {
	vec4 texColor = vec4(textColor, 1.0) * vec4(1.0, 1.0, 1.0, texture2D(texture, texCoord).r); 
	vec4 mixColor = mix(texColor, color, colorMixAmount);
	fragColor = vec4(mixColor.rgb * lighting, mixColor.a);
}