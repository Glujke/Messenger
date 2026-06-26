package state

import "fyne.io/fyne/v2"

// RunOnUI schedules fn on the Fyne main thread. Use from background goroutines only.
func RunOnUI(fn func()) {
	if fn == nil {
		return
	}
	fyne.Do(fn)
}
