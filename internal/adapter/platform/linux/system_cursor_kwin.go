//go:build linux

package linux

import (
	"context"
	"image"

	"github.com/godbus/dbus/v5"

	"github.com/y3owk1n/neru/internal/adapter/platform/kwin"
	"github.com/y3owk1n/neru/internal/ports"
)

const (
	kwinBusName          = "org.kde.KWin"
	kwinObjectPath       = "/KWin"
	kwinActiveOutputCall = "org.kde.KWin.activeOutputName"
)

// kwinActiveScreen picks the screen the overlay should go on. KWin can't tell
// us where the pointer is, so the cached cursor would stay on the first screen
// and hints on a second monitor would be dropped or drawn in the wrong place.
//
// In order: the screen of the focused window, the first screen when nothing is
// focused (the panel and desktop icons live there), then KWin's active output.
// That last one follows the pointer, so it only helps when the first two can't
// answer.
func kwinActiveScreen(ctx context.Context) (ports.Screen, bool) {
	screens, err := wlrootsScreens()
	if err != nil {
		return ports.Screen{}, false
	}

	window, focused, focusErr := kwin.Shared(nil).Focused()
	if focusErr == nil && focused {
		centre := window.Rect.Min.Add(window.Rect.Size().Div(2)) //nolint:mnd // centre

		for _, screen := range screens {
			if centre.In(screen.Bounds) {
				return screen, true
			}
		}
	}

	if focusErr == nil && !focused && len(screens) > 0 {
		return screens[0], true
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		return ports.Screen{}, false
	}

	var name string

	call := conn.Object(kwinBusName, kwinObjectPath).CallWithContext(ctx, kwinActiveOutputCall, 0)
	if call.Err != nil || call.Store(&name) != nil || name == "" {
		return ports.Screen{}, false
	}

	return ports.ScreenByName(screens, name)
}

// kwinCursorPosition keeps the cached cursor if it is already on the right
// screen, otherwise it moves it to the middle of that screen.
func kwinCursorPosition(ctx context.Context) (image.Point, bool) {
	screen, ok := kwinActiveScreen(ctx)
	if !ok {
		return image.Point{}, false
	}

	cached, err := waylandCursorPosition()
	if err == nil && cached.In(screen.Bounds) {
		return cached, true
	}

	return screen.Bounds.Min.Add(screen.Bounds.Size().Div(2)), true //nolint:mnd // centre
}
