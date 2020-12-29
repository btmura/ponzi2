package ui

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/logger"
)

// Internal constants.
var (
	// noSwappedIndices is the sentinel value returned by swapSidebarSlots if no swap happened.
	noSwappedIndices = [2]int{}

	// thumbSize is the size of the thumbnails in the sidebar.
	thumbSize = image.Pt(155, 105)

	// wheelScrollAmount is how much to scroll when the mouse scroll wheel is used.
	wheelScrollAmount = image.Pt(0, thumbSize.Y+viewPadding)

	// bumperScrollAmount is how much to scroll when dragging over the bumpers which are the edges of the sidebar.
	bumperScrollAmount = image.Pt(0, (thumbSize.Y+viewPadding)/4)
)

// sidebar manages the sidebar to display and edit stock thumbnails.
type sidebar struct {
	// priceStyle is the style to create thumbnails with.
	priceStyle chart.PriceStyle

	// slots are slots which can have thumbnails or be a drop site.
	slots []*sidebarSlot

	// draggedSlot is the slot being dragged if not nil.
	draggedSlot *draggedSidebarSlot

	// scrollOffset stores the Y offset accumulated from mouseScrollDirection event.
	scrollOffset int

	// bounds is the rectangle to draw within.
	bounds image.Rectangle

	// slotSwapCallback is a callback fired when the slots are swapped.
	slotSwapCallback func(i, j int)

	// thumbRemoveButtonClickCallback is called when a thumb's remove button is clicked.
	thumbRemoveButtonClickCallback func(symbol string)

	// thumbClickCallback is called when a thumb is clicked.
	thumbClickCallback func(symbol string)
}

// sidebarSlot is a slot in the sidebar that can contain thumbnails or be a drop site.
type sidebarSlot struct {
	// bounds is the rectangle to draw the slot in.
	bounds image.Rectangle

	// thumbs are the thumbnails in the slot.
	thumbs []*sidebarThumb

	// Fader fades out the slot.
	*view.Fader
}

// sidebarThumb wraps a thumbnail with a symbol and fader.
type sidebarThumb struct {
	// symbol is the symbol of the thumbnail.
	symbol string

	// Thumb is a stock thumbnail. Could have bounds outside the slot if dragged.
	*chart.Thumb

	// Fader fades out the thumbnail.
	*view.Fader
}

// draggedSidebarSlot wraps around a sidebarSlot with dragging information.
type draggedSidebarSlot struct {
	// sidebarSlot is the slot being dragged.
	*sidebarSlot

	// floatingBounds is the rectangle to draw within while the slot is being dragged.
	floatingBounds image.Rectangle

	// mousePressOffset is an offset from the center of the slot.
	mousePressOffset image.Point
}

func newSidebar() *sidebar {
	return new(sidebar)
}

func (s *sidebar) SetPriceStyle(newPriceStyle chart.PriceStyle) {
	if newPriceStyle == chart.PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}
	s.priceStyle = newPriceStyle
	for _, slot := range s.slots {
		for _, thumb := range slot.thumbs {
			thumb.SetPriceStyle(newPriceStyle)
		}
	}
}

func (s *sidebar) AddChartThumb(symbol string) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	thumb := chart.NewThumb(s.priceStyle)
	thumb.SetRemoveButtonClickCallback(func() {
		if s.thumbRemoveButtonClickCallback != nil {
			s.thumbRemoveButtonClickCallback(symbol)
		}
	})
	thumb.SetThumbClickCallback(func() {
		if s.thumbClickCallback != nil {
			s.thumbClickCallback(symbol)
		}
	})

	s.slots = append(s.slots, newSidebarSlot(symbol, thumb))
	return true
}

func (s *sidebar) RemoveChartThumb(symbol string) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	for _, slot := range s.slots {
		for _, thumb := range slot.thumbs {
			if symbol == thumb.symbol {
				changed = true
				slot.FadeOut()
			}
		}
	}
	return changed
}

func (s *sidebar) SetLoading(symbol string) (changed bool) {
	for _, slot := range s.slots {
		for _, thumb := range slot.thumbs {
			if symbol == thumb.symbol {
				changed = true
				thumb.SetLoading(true)
				thumb.SetErrorMessage("")
			}
		}
	}
	return changed
}

func (s *sidebar) SetData(symbol string, data chart.Data) (changed bool) {
	for _, slot := range s.slots {
		for _, thumb := range slot.thumbs {
			if symbol == thumb.symbol {
				changed = true
				thumb.SetLoading(false)
				thumb.SetData(data)
			}
		}
	}
	return changed
}

func (s *sidebar) SetErrorMessage(symbol string, errorMessage string) (changed bool) {
	for _, slot := range s.slots {
		for _, thumb := range slot.thumbs {
			if symbol == thumb.symbol {
				changed = true
				thumb.SetLoading(false)
				thumb.SetErrorMessage(errorMessage)
			}
		}
	}
	return changed
}

// ContentSize returns the size of the sidebar's contents like thumbnails
// which could be less than the sidebar's bounds if there are not that many thumbnails.
func (s *sidebar) ContentSize() image.Point {
	num := len(s.slots)
	if num == 0 {
		return image.Pt(0, 0)
	}

	height := num * thumbSize.Y
	if num > 1 {
		// Add padding between thumbnails.
		height += (num - 1) * viewPadding

		// Add padding on top and bottom.
		height += 2 * viewPadding
	}
	return image.Pt(thumbSize.X, height)
}

func (s *sidebar) SetBounds(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *sidebar) ProcessInput(input *view.Input) {
	if input == nil {
		logger.Error("input should not be nil")
		return
	}

	// Adjust the scroll offset due to mouse wheel scrolling or dragging over the bumpers.
	s.adjustScrollOffset(input)

	// Set each slot's bounds on the screen.
	s.setSlotBounds()

	// Find the slot being dragged and update its position.
	wasDragging := s.draggedSlot != nil
	s.setDraggedSlot(input)
	swappedIndices := s.swapDraggedSlot(input)
	stillDragging := s.draggedSlot != nil

	// Absorb mouse event if dragging was released or still dragging.
	if wasDragging && !stillDragging || stillDragging {
		input.ClearMouseInput()
	}

	// Forward the input to the individual slots.
	for _, slot := range s.slots {
		slot.ProcessInput(input)
	}

	// Schedule a change event if slots were swapped.
	if swappedIndices != noSwappedIndices {
		s.fireSwapCallback(input, swappedIndices)
	}
}

func (s *sidebar) adjustScrollOffset(input *view.Input) {
	if input == nil {
		logger.Error("input should not be nil")
		return
	}

	// Adjust the mouseScrollDirection offset due to the mouse wheel or window size.
	h := s.ContentSize().Y
	if h == 0 || h < s.bounds.Dy() {
		// Reset offset if the sidebar is empty or the contents are less than the bounds.
		s.scrollOffset = 0
		return
	}

	// Scroll up or down if dragging and hovering over the bumpers.
	if input.MouseLeftButtonDragging != nil && s.draggedSlot != nil {
		pos := input.MouseLeftButtonDragging.CurrentPos
		top := image.Rect(
			s.bounds.Min.X, s.bounds.Max.Y-thumbSize.Y/2,
			s.bounds.Max.X, s.bounds.Max.Y)
		bottom := image.Rect(
			s.bounds.Min.X, 0,
			s.bounds.Max.X, thumbSize.Y/2)
		switch {
		case pos.In(top):
			s.scrollUp(bumperScrollAmount.Y)
			return
		case pos.In(bottom):
			s.scrollDown(bumperScrollAmount.Y)
			return
		}
	}

	// Scroll up or down if the mouseScrollDirection event occurred within the sidebar bounds.
	if input.MouseScrolled.In(s.bounds) {
		switch input.MouseScrolled.Direction {
		case view.ScrollUp:
			s.scrollUp(wheelScrollAmount.Y)
			return
		case view.ScrollDown:
			s.scrollDown(wheelScrollAmount.Y)
			return
		default:
			logger.Errorf("unsupported scroll direction: %v", input.MouseScrolled.Direction)
			return
		}
	}
}

func (s *sidebar) scrollUp(scrollAmount int) {
	// Don't allow a gap at the top.
	s.scrollOffset += scrollAmount
	if s.scrollOffset > 0 {
		s.scrollOffset = 0
	}
}

func (s *sidebar) scrollDown(scrollAmount int) {
	// Don't scroll if there is no need.
	overflow := s.ContentSize().Y - s.bounds.Dy()
	if overflow <= 0 {
		return
	}

	// Don't allow a gap at the bottom.
	scroll := scrollAmount
	if available := s.scrollOffset + overflow; scroll > available {
		scroll = available
	}
	s.scrollOffset -= scroll
}

// setSlotBounds goes through the sidebar and assign bounds to each slot.
func (s *sidebar) setSlotBounds() {
	slotBounds := image.Rect(
		s.bounds.Min.X, s.bounds.Max.Y-viewPadding-thumbSize.Y,
		s.bounds.Max.X, s.bounds.Max.Y-viewPadding,
	)
	slotBounds = slotBounds.Sub(image.Pt(0, s.scrollOffset))

	for _, slot := range s.slots {
		slot.bounds = slotBounds
		for _, thumb := range slot.thumbs {
			thumb.SetBounds(slotBounds)
		}
		slotBounds = slotBounds.Sub(image.Pt(0, thumbSize.Y+viewPadding))
	}
}

// setDraggedSlot finds the slot being dragged and updates its position.
func (s *sidebar) setDraggedSlot(input *view.Input) {
	if input == nil {
		logger.Error("input should not be nil")
		return
	}

	if input.MouseLeftButtonDragging == nil {
		s.draggedSlot = nil
	}

	for _, slot := range s.slots {
		if s.draggedSlot == nil && input.MouseLeftButtonDragging.PressedIn(slot.bounds) {
			s.draggedSlot = &draggedSidebarSlot{
				sidebarSlot:      slot,
				mousePressOffset: input.MouseLeftButtonDragging.CurrentPos.Sub(rect.CenterPoint(slot.bounds)),
			}
		}
	}
}

// swapDraggedSlot swaps the dragged slot up or down depending on the mouse position.
func (s *sidebar) swapDraggedSlot(input *view.Input) (swappedIndexes [2]int) {
	if input == nil {
		logger.Error("input should not be nil")
		return noSwappedIndices
	}

	if s.draggedSlot != nil && input.MouseLeftButtonDragging == nil {
		logger.Error("should be dragging if dragged slot exists")
		return noSwappedIndices
	}

	if s.draggedSlot == nil {
		return noSwappedIndices
	}

	// Determine the center from where the user pressed and held.
	center := input.MouseLeftButtonDragging.CurrentPos.Sub(s.draggedSlot.mousePressOffset)

	// Float the dragged slot to be under the mouse cursor.
	thumbBounds := rect.FromCenterPointAndSize(center, thumbSize)
	s.draggedSlot.floatingBounds = thumbBounds
	for _, thumb := range s.draggedSlot.thumbs {
		thumb.SetBounds(thumbBounds)
	}

	// Determine whether to move the dragged slot up or down
	// by checking whether the mouse is moving up or down.

	currentPos := input.MouseLeftButtonDragging.CurrentPos
	previousPos := input.MouseLeftButtonDragging.PreviousPos

	var dy int
	switch {
	case currentPos.Y < previousPos.Y:
		dy = +1
	case currentPos.Y > previousPos.Y:
		dy = -1
	}

	// If the mouse is not moving, then we don't need to do anything.
	if dy == 0 {
		return noSwappedIndices
	}

	// Find the index of the dragged slot.
	draggedSlotIndex := -1
	for i := range s.slots {
		if s.draggedSlot.sidebarSlot == s.slots[i] {
			draggedSlotIndex = i
			break
		}
	}

	if draggedSlotIndex < 0 {
		logger.Error("should find dragged slot if draggedSlot exists")
		return noSwappedIndices
	}

	// Find the slot to swap with and swap it.
	for i := draggedSlotIndex + dy; i >= 0 && i < len(s.slots); i += dy {
		b := s.slots[i].bounds
		xOverlaps := b.Overlaps(s.draggedSlot.floatingBounds)
		yOverlaps := b.Min.Y <= currentPos.Y && currentPos.Y < b.Max.Y
		if xOverlaps && yOverlaps {
			j, k := draggedSlotIndex, i
			s.slots[j], s.slots[k] = s.slots[k], s.slots[j]
			return [2]int{draggedSlotIndex, i}
		}
	}

	return noSwappedIndices
}

// fireSwapCallback schedules the swap callback.
func (s *sidebar) fireSwapCallback(input *view.Input, swappedIndices [2]int) {
	if input == nil {
		logger.Error("input should not be nil")
		return
	}

	if swappedIndices == noSwappedIndices {
		logger.Error("swappedIndices should not be the zero value")
		return
	}

	if s.slotSwapCallback == nil {
		return
	}

	input.AddFiredCallback(func() {
		s.slotSwapCallback(swappedIndices[0], swappedIndices[1])
	})
}

// Update moves the animation one step forward.
func (s *sidebar) Update() (dirty bool) {
	for i := 0; i < len(s.slots); i++ {
		slot := s.slots[i]
		if slot.Update() {
			dirty = true
		}
		if slot.DoneFadingOut() {
			s.slots = append(s.slots[:i], s.slots[i+1:]...)
			slot.Close()
			i--
		}
	}
	return dirty
}

// Render renders a frame.
func (s *sidebar) Render(fudge float32) {
	// Draw the non-dragged thumbnails first, so they appear under the dragged thumbnail.
	for _, slot := range s.slots {
		if s.draggedSlot != nil && s.draggedSlot.sidebarSlot == slot {
			continue
		}
		slot.Render(fudge)
	}

	if s.draggedSlot != nil {
		s.draggedSlot.Render(fudge)
	}
}

func (s *sidebar) SetSlotSwapCallback(cb func(i, j int)) {
	s.slotSwapCallback = cb
}

func (s *sidebar) SetThumbRemoveButtonClickCallback(cb func(symbol string)) {
	s.thumbRemoveButtonClickCallback = cb
}

func (s *sidebar) SetThumbClickCallback(cb func(symbol string)) {
	s.thumbClickCallback = cb
}

func (s *sidebar) Close() {
	s.slotSwapCallback = nil
	s.thumbRemoveButtonClickCallback = nil
	s.thumbClickCallback = nil
}

func newSidebarSlot(symbol string, thumb *chart.Thumb) *sidebarSlot {
	return &sidebarSlot{
		thumbs: []*sidebarThumb{newSidebarThumb(symbol, thumb)},
		Fader:  view.NewStartedFader(1 * view.FPS),
	}
}

func (s *sidebarSlot) ProcessInput(input *view.Input) {
	if input == nil {
		logger.Error("input should not be nil")
		return
	}

	for _, thumb := range s.thumbs {
		thumb.ProcessInput(input)
	}
}

func (s *sidebarSlot) Update() (dirty bool) {
	for _, thumb := range s.thumbs {
		if thumb.Update() {
			dirty = true
		}
	}
	if s.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (s *sidebarSlot) Render(fudge float32) {
	s.Fader.Render(fudge, func() {
		for _, thumb := range s.thumbs {
			thumb.Render(fudge)
		}
	})
}

func (s *sidebarSlot) Close() {
	for _, thumb := range s.thumbs {
		thumb.Close()
	}
	s.thumbs = nil
}

func newSidebarThumb(symbol string, thumb *chart.Thumb) *sidebarThumb {
	return &sidebarThumb{
		symbol: symbol,
		Thumb:  thumb,
		Fader:  view.NewStartedFader(1 * view.FPS),
	}
}

func (s *sidebarThumb) Update() (dirty bool) {
	if s.Thumb.Update() {
		dirty = true
	}
	if s.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (s *sidebarThumb) Render(fudge float32) {
	s.Fader.Render(fudge, func() {
		s.Thumb.Render(fudge)
	})
}
