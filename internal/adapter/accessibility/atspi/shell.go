//go:build linux

package atspi

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/adapter/accessibility/ax"
	"github.com/y3owk1n/neru/internal/adapter/platform/kwin"
)

// The KDE panel, tray and desktop icons all belong to plasmashell, which AT-SPI
// lists as its own application. Hints only walk the focused window, so these
// are added when the app is named in hints.additional_menubar_hints_targets.

// Plasmashell reports a desktop icon as a canvas. We hand it on as a list item
// because the role filter afterwards only knows the configured roles.
const (
	roleCanvas   = "canvas"
	roleListItem = "list item"
)

// SupportsSupplementaryElements lets the adapter scan the configured extra apps.
func (c *Client) SupportsSupplementaryElements() bool { return true }

// MenuBarClickableElements returns nothing. Linux has no global menu bar, but
// the adapter asks for it first and an error would log on every activation.
func (c *Client) MenuBarClickableElements(context.Context, int) ([]ax.Node, error) {
	return nil, nil
}

// ClickableElementsFromBundleID scans every AT-SPI app called appName.
// Plasmashell reports real screen coordinates, so no window origin is added.
func (c *Client) ClickableElementsFromBundleID(
	ctx context.Context,
	appName string,
	roles []string,
	maxDepth int,
) ([]ax.Node, error) {
	conn, err := c.ensureA11yConn()
	if err != nil {
		return nil, err
	}

	var (
		out    []ax.Node
		wanted = shellRoles(roles, desktopFocused())
		depth  = c.treeDepth(maxDepth)
		screen = placement{coordType: atspiCoordScreen}
	)

	for _, app := range c.children(ctx, conn, accRef{Name: atspiRegistryDest, Path: atspiRootPath}) {
		if !strings.EqualFold(strings.TrimSpace(c.name(ctx, conn, app)), appName) {
			continue
		}

		nodes, path, visited := c.scan(ctx, conn, app, wanted, false, depth, screen, false)

		c.logger.Debug("AT-SPI extra app scan",
			zap.String("app", appName),
			zap.String("path", path),
			zap.Int("count", len(nodes)),
			zap.Int("visited", visited))

		for _, node := range nodes {
			if icon, ok := node.(*atspiNode); ok && icon.role == roleCanvas {
				icon.role = roleListItem
			}
		}

		out = append(out, nodes...)
	}

	return out, nil
}

// desktopFocused is true when KWin says no window has focus, i.e. the desktop
// is in front. Desktop icons are scanned even when a window covers them, so
// they are only included in this case.
func desktopFocused() bool {
	_, found, err := kwin.Shared(nil).Focused()

	return err == nil && !found
}

// shellRoles is the default clickable roles plus list items and whatever the
// caller passed. The caller's list is the menu bar's, which has no buttons on
// Linux, so on its own it would match nothing in a panel. Canvas is too common
// elsewhere to hint by default, so it is only added for the desktop icons.
func shellRoles(extra []string, withDesktopIcons bool) map[string]struct{} {
	set := map[string]struct{}{roleListItem: {}}

	if withDesktopIcons {
		set[roleCanvas] = struct{}{}
	}

	for role := range defaultClickableRoles {
		set[role] = struct{}{}
	}

	for _, role := range extra {
		if trimmed := strings.ToLower(strings.TrimSpace(role)); trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}

	return set
}
