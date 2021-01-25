package main

import (
	"fmt"
	"os"

	"ds-to-dhall/dhall2ds"
	"ds-to-dhall/dockerimg"
	"ds-to-dhall/ds2dhall"
	"github.com/inconshreveable/log15"
)

func main() {
	cmds := make(map[string]func([]string))
	cmds["ds2dhall"] = ds2dhall.Main
	cmds["dockerimg"] = dockerimg.Main
	cmds["dhall2ds"] = dhall2ds.Main

	log15.Root().SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))

	if len(os.Args) < 2 {
		fmt.Println("expected a subcommands")
		os.Exit(1)
	}

	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Printf("unknown subcommand %s", os.Args[1])
		os.Exit(1)
	}

	cmd(os.Args[2:])
}
