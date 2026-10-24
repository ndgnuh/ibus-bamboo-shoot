/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) 2018 Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BambooEngine/bamboo-core"
	ibus "github.com/BambooEngine/goibus"
	"ibus-bamboo/config"
	"ibus-bamboo/introspectors"
)

const (
	ComponentName = "org.freedesktop.IBus.bamboo"
	EngineName    = "Bamboo"
)

var embedded = flag.Bool("ibus", false, "Run the embedded ibus component")
var version = flag.Bool("version", false, "Show version")
var gui = flag.Bool("gui", false, "Show GUI")
var isWayland = false
var isGnome = false
var engineRef *IBusBambooEngine
var introspector introspectors.BambooIntrospector

func hasGnome(env string) bool {
	return strings.Contains(strings.ToLower(os.Getenv(env)), "gnome")
}

// guessIntrospector picks an introspector from the running session.
// GNOME is checked first: GNOME on Wayland does not implement the
// wlr foreign toplevel protocol used by the Wayland introspector.
func guessIntrospector() introspectors.BambooIntrospector {
	if isGnome {
		return introspectors.IntrospectorGnome
	}
	if isWayland {
		return introspectors.IntrospectorWayland
	}
	return introspectors.IntrospectorX11
}

func validateConfig(cfg *config.Config) error {
	resolved, ok := introspectors.IntrospectorByName[string(cfg.Introspector)]
	if !ok {
		return fmt.Errorf("unknown introspector %q", cfg.Introspector)
	}
	cfg.Introspector = resolved
	if cfg.Introspector == introspectors.IntrospectorCommand && cfg.IntrospectorCommand == "" {
		return fmt.Errorf("introspector %q requires IntrospectorCommand", cfg.Introspector)
	}
	if _, ok := cfg.InputMethodDefinitions[cfg.InputMethod]; !ok {
		return fmt.Errorf("unknown input method %q", cfg.InputMethod)
	}
	for _, cs := range bamboo.GetCharsetNames() {
		if cs == cfg.OutputCharset {
			return nil
		}
	}
	return fmt.Errorf("unknown output charset %q", cfg.OutputCharset)
}

func main() {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		isWayland = true
	}
	if hasGnome("XDG_CURRENT_DESKTOP") || hasGnome("DESKTOP_SESSION") || hasGnome("GDMSESSION") {
		isGnome = true
	}

	flag.Parse()

	/// Print the version and exit
	if *version {
		fmt.Println(Version)
		return
	}

	if *embedded {
		os.Chdir(DataDir)
	}

	// This should not be spawned here, now that we have multiple introspectors
	// if isWayland && !isGnome {
	// 	go introspectors.WlGetFocusWindowClass()
	// }

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg := config.LoadConfig(strings.ToLower(EngineName))
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Invalid config %s: %v", config.GetConfigPath(strings.ToLower(EngineName)), err)
	}
	if cfg.Introspector == introspectors.IntrospectorAuto {
		cfg.Introspector = guessIntrospector()
	}
	introspector = cfg.Introspector
	log.Printf("Using introspector: %s", introspector)
	introspectors.BootstrapIntrospector(introspector, &cfg.IntrospectorCommand)

	if *embedded {
		log.Println("IbusBamboo running in embedded mode")
		engine := GetIBusEngineCreator()
		bus := ibus.NewBus()
		bus.RequestName(ComponentName, 0)

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, engine)

		select {}
	} else {
		log.Println("IbusBamboo running in IBus mode")
		bus := ibus.NewBus()
		log.Println("Got Bus, Running Standalone")
		component := &ibus.Component{
			Name:          "IBusComponent",
			ComponentName: ComponentName + "Standalone",
		}
		engine := &ibus.EngineDesc{
			Name:       "IBusEngineDesc",
			EngineName: EngineName + "Standalone",
		}
		component.AddEngine(engine)
		bus.RegisterComponent(component)

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, GetIBusEngineCreator())

		bus.CallMethod("SetGlobalEngine", 0, EngineName+"Standalone")

		select {}
	}
}
