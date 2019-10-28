#version 450 core

#define FRAG_COLOR_MODE 0
#define FRAG_TEXTURE_MODE 1
#define FRAG_TEXT_COLOR_MODE 2

layout(location = 5) uniform int fragMode;
layout(location = 6) uniform sampler2D texture;
layout(location = 7) uniform vec4 textColor;
layout(location = 8) uniform float alpha;

in vec4 color;
in vec2 texCoord;

out vec4 fragColor;

void main(void) {
	switch (fragMode) {
	case FRAG_COLOR_MODE:
		fragColor = color;
		break;

	case FRAG_TEXTURE_MODE:
		fragColor = texture2D(texture, texCoord);
		break;

	case FRAG_TEXT_COLOR_MODE:
		fragColor = vec4(textColor.rgb, texture2D(texture, texCoord).r);
		break;
	}

	fragColor = vec4(fragColor.rgb, fragColor.a * alpha);
}