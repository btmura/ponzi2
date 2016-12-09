#version 450 core

layout(location = 9) uniform sampler2D texture;

in vec2 texCoord;
in vec3 lighting;

out vec4 fragColor;

void main(void) {
	vec4 color = texture2D(texture, texCoord);
	fragColor = vec4(color.rgb * lighting, color.a);
}