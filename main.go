package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"ds-to-dhall/dhall2ds"
	"ds-to-dhall/dockerimg"
	"ds-to-dhall/ds2dhall"
	"github.com/inconshreveable/log15"
)

func main() {
	cmds := make(map[string]func([]string))
	shortDescriptions := make(map[string]string)
	cmds["ds2dhall"] = ds2dhall.Main
	shortDescriptions["ds2dhall"] = ds2dhall.ShortDescription
	cmds["dockerimg"] = dockerimg.Main
	shortDescriptions["dockerimg"] = dockerimg.ShortDescription
	cmds["dhall2ds"] = dhall2ds.Main
	shortDescriptions["dhall2ds"] = dhall2ds.ShortDescription

	cmdNames := make([]string, 0, len(cmds)+1)
	for cmdName := range cmds {
		cmdNames = append(cmdNames, cmdName)
	}
	cmdNames = append(cmdNames, "help")

	log15.Root().SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))

	if len(os.Args) < 2 {
		fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
		os.Exit(1)
	}

	if os.Args[1] == "help" {
		if len(os.Args) == 2 {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

			for cmdName := range cmds {
				fmt.Fprintf(w, "\t%s\t%s\n", cmdName, shortDescriptions[cmdName])
			}
			w.Flush()
			os.Exit(0)
		}

		cmd, ok := cmds[os.Args[2]]
		if !ok {
			fmt.Printf("unknown subcommand %s\n", os.Args[1])
			fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
			os.Exit(1)
		}

		cmd([]string{"-h"})
	}

	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Printf("unknown subcommand %s\n", os.Args[1])
		fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
		os.Exit(1)
	}

	cmd(os.Args[2:])
}
