package color

// Colors used throughout the UI.
var (
	Green           = RGBA{0.25, 1, 0, 1}
	Red             = RGBA{1, 0.3, 0, 1}
	Yellow          = RGBA{1, 1, 0, 1}
	Purple          = RGBA{0.5, 0, 1, 1}
	White           = RGBA{1, 1, 1, 1}
	Gray            = RGBA{0.15, 0.15, 0.15, 1}
	TransparentGray = RGBA{0.15, 0.15, 0.15, 0.5}
	LightGray       = RGBA{0.35, 0.35, 0.35, 1}
	Orange          = RGBA{1, 0.5, 0, 1}
)

// RGBA is color with red, green, blue, and alpha components.
type RGBA [4]float32
