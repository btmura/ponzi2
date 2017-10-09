#version 450 core

layout(location = 5) uniform int mode;
layout(location = 6) uniform sampler2D texture;
layout(location = 7) uniform vec3 textColor;

in vec4 color;
in vec2 texCoord;

out vec4 fragColor;

void main(void) {
	switch (mode) {
	case 0:
		fragColor = color;
		break;

	case 1:
		fragColor = texture2D(texture, texCoord);
		break;

	case 2:
		fragColor = vec4(textColor, texture2D(texture, texCoord).r);
		break;
	}
}