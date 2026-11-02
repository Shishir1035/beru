//go:build linux

package atspi

import (
	"context"

	"github.com/y3owk1n/neru/internal/adapter/accessibility/ax"
	"github.com/y3owk1n/neru/internal/adapter/platform"
	"github.com/y3owk1n/neru/internal/adapter/platform/kwin"
	"github.com/y3owk1n/neru/internal/adapter/platform/linux"
	"github.com/y3owk1n/neru/internal/derrors"
)

// focusedApp is the focused app's app_id or WM_CLASS, which is what
// excluded_apps and per-app config match against.
type focusedApp struct{ appID string }

func (a focusedApp) Release() {}

func (a focusedApp) BundleIdentifier() string { return a.appID }

func (a focusedApp) Info() (*ax.AppInfo, error) {
	return &ax.AppInfo{Role: "application", Title: a.appID}, nil
}

// FocusedApplication asks the compositor which app has focus. The embedded
// client always failed here on Linux, so excluded_apps never matched. KWin has
// no toplevel protocol, so on KDE the name is the window class from its script.
func (c *Client) FocusedApplication(context.Context) (ax.App, error) {
	backend := platform.DetectLinuxBackend()

	var appID string

	if backend == platform.BackendWaylandKDE {
		if window, found, err := kwin.Shared(nil).Focused(); err == nil && found {
			appID = window.Class
		}
	} else {
		appID, _, _ = linux.FocusedAppIdentity(backend.String())
	}

	if appID == "" {
		return nil, derrors.New(derrors.CodeAccessibilityFailed, "no focused application")
	}

	return focusedApp{appID: appID}, nil
}
