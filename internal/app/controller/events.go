package controller

import (
	"context"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
)

func (c *Controller) setEventCallbacks(ctx context.Context, win *glfw.Window) {
	win.SetCharCallback(func(_ *glfw.Window, char rune) {
		glog.V(2).Infof("char: %c", char)
		c.setChar(char)
	})

	win.SetCharModsCallback(func(_ *glfw.Window, char rune, mods glfw.ModifierKey) {
		glog.V(2).Infof("char mods: %c, %v", char, mods)
	})

	win.SetCloseCallback(func(_ *glfw.Window) {
		glog.V(2).Infof("close")
	})

	win.SetCursorEnterCallback(func(_ *glfw.Window, entered bool) {
		glog.V(2).Infof("cursor enter: %t", entered)
	})

	win.SetCursorPosCallback(func(_ *glfw.Window, xpos, ypos float64) {
		glog.V(2).Infof("cursor pos: %f, %f", xpos, ypos)
		c.setCursorPos(xpos, ypos)
	})

	win.SetDropCallback(func(_ *glfw.Window, names []string) {
		glog.V(2).Infof("drop: %v", names)
	})

	win.SetFocusCallback(func(_ *glfw.Window, focused bool) {
		glog.V(2).Infof("focus: %t", focused)
	})

	win.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		glog.V(2).Infof("framebuffer size: %d, %d", width, height)
	})

	win.SetIconifyCallback(func(_ *glfw.Window, iconified bool) {
		glog.V(2).Infof("iconify: %t", iconified)
	})

	win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		glog.V(2).Infof("key: %v, %d, %v, %v", key, scancode, action, mods)
		c.setKey(ctx, key, action)
	})

	win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		glog.V(2).Infof("mouse button: %v, %v", button, action, mods)
		c.setMouseButton(button, action)
	})

	win.SetPosCallback(func(_ *glfw.Window, xpos, ypos int) {
		glog.V(2).Infof("pos: %d, %d", xpos, ypos)
	})

	win.SetRefreshCallback(func(_ *glfw.Window) {
		glog.V(2).Infof("refresh")
	})

	win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		glog.V(2).Infof("scroll: %f, %f", xoff, yoff)
	})

	win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		glog.V(2).Infof("size: %d, %d", width, height)
		c.setSize(width, height)
	})
}
