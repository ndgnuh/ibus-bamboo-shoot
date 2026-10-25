package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"ibus-bamboo/config"
	"ibus-bamboo/introspectors"
)

var modeCommands []BambooCliCommand

// parseMode accepts either the integer value or the short name of a typing mode
func parseMode(arg string) (int, error) {
	if mode, err := strconv.Atoi(arg); err == nil {
		if _, ok := config.ImLookupTable[mode]; ok {
			return mode, nil
		}
	}
	for mode, name := range config.ImShortNames {
		if name == arg {
			return mode, nil
		}
	}
	return 0, fmt.Errorf("unknown typing mode %q", arg)
}

var modeHelpCommand = BambooCliCommand{
	name: "help",
	help: "Show help for typing mode commands",
	dispatch: func(args []string) error {
		printHelp("bamboo-cli mode - Manage the typing mode", modeCommands)
		return nil
	},
}

var modeListCommand = BambooCliCommand{
	name: "list",
	help: "List all typing modes",
	dispatch: func(args []string) error {
		var modes []int
		for mode := range config.ImLookupTable {
			modes = append(modes, mode)
		}
		sort.Ints(modes)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		for _, mode := range modes {
			fmt.Fprintf(w, "%d\t%s\t%s\n", mode, config.ImShortNames[mode], config.ImLookupTable[mode])
		}
		w.Flush()
		return nil
	},
}

var modeGetCommand = BambooCliCommand{
	name: "get",
	help: "Get current typing mode (for focused client)",
	dispatch: func(args []string) error {
		cfg := config.LoadConfig(engineName)
		introspectors.BootstrapIntrospector(cfg)
		wmClass := introspectors.GetLatestAppId(cfg.Introspector)

		// Same resolution as the engine's getInputMode
		mode := config.PreeditIM
		if im, ok := cfg.InputModeMapping[wmClass]; ok && wmClass != "" && config.ImLookupTable[im] != "" {
			mode = im
		} else if _, ok := config.ImLookupTable[cfg.DefaultInputMode]; ok {
			mode = cfg.DefaultInputMode
		}
		fmt.Println(config.ImShortNames[mode])
		return nil
	},
}

var modeGetDefaultCommand = BambooCliCommand{
	name: "get-default",
	help: "Print current default typing mode",
	dispatch: func(args []string) error {
		cfg := config.LoadConfig(engineName)
		fmt.Println(config.ImShortNames[cfg.DefaultInputMode])
		return nil
	},
}

var modeSetDefaultCommand = BambooCliCommand{
	name: "set-default",
	help: "Set default typing mode (integer or short name)",
	dispatch: func(args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("usage: mode set-default <mode>")
		}
		mode, err := parseMode(args[0])
		if err != nil {
			return err
		}
		cfg := config.LoadConfig(engineName)
		cfg.DefaultInputMode = mode
		config.SaveConfig(cfg, engineName)
		fmt.Println(config.ImShortNames[mode])
		return nil
	},
}

var modeSetCommand = BambooCliCommand{
	name: "set",
	help: "Set typing mode for an app id/wm_class (integer or short name)",
	dispatch: func(args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("usage: mode set <app_id/wm_class> <mode>")
		}
		mode, err := parseMode(args[1])
		if err != nil {
			return err
		}
		cfg := config.LoadConfig(engineName)
		cfg.InputModeMapping[args[0]] = mode
		config.SaveConfig(cfg, engineName)
		fmt.Println(config.ImShortNames[mode])
		return nil
	},
}

var modeCommand = BambooCliCommand{
	name: "mode",
	help: "Manage the typing mode",
	dispatch: func(args []string) error {
		if len(args) == 0 {
			modeHelpCommand.dispatch(args)
			return fmt.Errorf("missing mode command")
		}

		var action, rest = args[0], args[1:]
		for _, command := range modeCommands {
			if command.name == action {
				return command.dispatch(rest)
			}
		}

		modeHelpCommand.dispatch(rest)
		return fmt.Errorf("unknown mode command %s", action)
	},
}

// Assigned in init to avoid an initialization cycle with modeHelpCommand
func init() {
	modeCommands = []BambooCliCommand{
		modeHelpCommand,
		modeListCommand,
		modeGetCommand,
		modeGetDefaultCommand,
		modeSetDefaultCommand,
		modeSetCommand,
	}
}
