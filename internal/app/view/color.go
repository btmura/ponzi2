package view

// Colors used throughout the UI.
var (
	Green                = Color{0.25, 1, 0, 1}
	Red                  = Color{1, 0.3, 0, 1}
	Yellow               = Color{1, 1, 0, 1}
	Purple               = Color{0.5, 0, 1, 1}
	White                = Color{1, 1, 1, 1}
	Gray                 = Color{0.15, 0.15, 0.15, 1}
	TransparentGray      = Color{0.15, 0.15, 0.15, 0.5}
	LightGray            = Color{0.35, 0.35, 0.35, 1}
	TransparentLightGray = Color{0.15, 0.15, 0.15, 0.5}
	Orange               = Color{1, 0.5, 0, 1}
)

// Color is color with red, green, blue, and alpha components.
type Color [4]float32
