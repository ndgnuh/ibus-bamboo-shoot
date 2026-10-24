package introspectors

import "strings"

type BambooIntrospector string
type IntrospectorMode int

const (
	// The CurrentClientId state is stored in this module
	// A subprocess will be spawn to watch over the currently
	// active client
	Reactive IntrospectorMode = iota

	// The current client id will be retreived by IBus Bamboo
	// when it is needed, by some kind of API. No subprocess is needed for this one.
	Passive
)

const (
	IntrospectorAuto    BambooIntrospector = "auto"
	IntrospectorX11     BambooIntrospector = "x11"
	IntrospectorGnome   BambooIntrospector = "gnome"
	IntrospectorWayland BambooIntrospector = "wayland"
	IntrospectorCommand BambooIntrospector = "command"
)

var IntrospectorNames = map[BambooIntrospector]string{
	IntrospectorAuto:    "auto",
	IntrospectorX11:     "x11",
	IntrospectorGnome:   "gnome",
	IntrospectorWayland: "wayland",
	IntrospectorCommand: "command",
}

var IntrospectorByName = map[string]BambooIntrospector{
	"":        IntrospectorAuto,
	"auto":    IntrospectorAuto,
	"x11":     IntrospectorX11,
	"gnome":   IntrospectorGnome,
	"wayland": IntrospectorWayland,
	"command": IntrospectorCommand,
}

// GetIntrospectorMode tells whether an introspector keeps the focused window
// class up to date in the background (Reactive) or queries it on demand (Passive).
func GetIntrospectorMode(introspector BambooIntrospector) IntrospectorMode {
	switch introspector {
	case IntrospectorWayland:
		// WlGetFocusWindowClass runs in a goroutine and keeps WlAppId updated
		return Reactive
	case IntrospectorX11, IntrospectorGnome, IntrospectorCommand:
		return Passive
	default:
		// IntrospectorAuto currently falls back to querying X11 on demand
		return Passive
	}
}

// BootstrapIntrospector starts whatever background work an introspector needs
// before the focused window class can be read from it.
// cmd is the command run by IntrospectorCommand, nil if not configured.
func BootstrapIntrospector(introspector BambooIntrospector, cmd *string) {
	switch introspector {
	case IntrospectorWayland:
		go WlGetFocusWindowClass()
	case IntrospectorCommand:
		if cmd != nil {
			CmdSetIntrospectCommand(*cmd)
		}
	}
}

// GetLatestAppId returns the class of the currently focused window,
// falling back to X11 when the introspector cannot tell.
func GetLatestAppId(introspector BambooIntrospector) string {
	var wmClass string
	switch introspector {
	case IntrospectorGnome:
		wmClass, _ = GnomeGetFocusWindowClass()
	case IntrospectorWayland:
		wmClass = WlAppId
	case IntrospectorX11:
		wmClass = X11GetFocusWindowClass()
	case IntrospectorCommand:
		wmClass = CmdGetFocusedWmClass()
	}

	if wmClass == "" {
		wmClass = X11GetFocusWindowClass()
	}
	wmClass = strings.Replace(wmClass, "\"", "", -1)
	return wmClass
}
