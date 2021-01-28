package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"ds-to-dhall/dhall2ds"
	"ds-to-dhall/dockerimg"
	"ds-to-dhall/ds2dhall"
	"github.com/inconshreveable/log15"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func versionString(version, commit, date string) string {
	b := bytes.Buffer{}
	w := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	fmt.Fprintf(w, "version:\t%s", version)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "commit:\t%s", commit)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "build date:\t%s", date)
	w.Flush()

	return b.String()
}

func main() {
	cmds := make(map[string]func([]string, context.Context))
	shortDescriptions := make(map[string]string)
	cmds["ds2dhall"] = ds2dhall.Main
	shortDescriptions["ds2dhall"] = ds2dhall.ShortDescription
	cmds["dockerimg"] = dockerimg.Main
	shortDescriptions["dockerimg"] = dockerimg.ShortDescription
	cmds["dhall2ds"] = dhall2ds.Main
	shortDescriptions["dhall2ds"] = dhall2ds.ShortDescription

	cmdNames := make([]string, 0, len(cmds)+2)
	for cmdName := range cmds {
		cmdNames = append(cmdNames, cmdName)
	}
	cmdNames = append(cmdNames, "help")
	cmdNames = append(cmdNames, "version")

	log15.Root().SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))

	if len(os.Args) < 2 {
		fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
		os.Exit(1)
	}

	if os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		if len(os.Args) == 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Println("ds-to-dhall <command>")
			fmt.Println("available commands:")
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)

			for cmdName := range cmds {
				fmt.Fprintf(w, "\t%s\t%s\n", cmdName, shortDescriptions[cmdName])
			}
			fmt.Fprintln(w, "\thelp\tshows help for commands")
			fmt.Fprintln(w, "\tversion\tshows version string")
			w.Flush()
			os.Exit(0)
		}

		cmd, ok := cmds[os.Args[2]]
		if !ok {
			fmt.Printf("unknown subcommand %s\n", os.Args[1])
			fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
			os.Exit(1)
		}

		cmd([]string{"-h"}, context.Background())
	}

	if os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v" {
		output := versionString(version, commit, date)
		fmt.Fprintln(os.Stderr, output)
		os.Exit(0)
	}

	cmd, ok := cmds[os.Args[1]]
	if !ok {
		fmt.Printf("unknown subcommand %s\n", os.Args[1])
		fmt.Printf("expected a subcommand: %s\n", strings.Join(cmdNames, ", "))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdown := func() {
		cancel()
	}
	defer shutdown()

	go trapSignalsForShutdown(shutdown)

	cmd(os.Args[2:], ctx)
}

func trapSignalsForShutdown(shutdown func()) {
	// Listen for shutdown signals. When we receive one attempt to clean up,
	// but do an insta-shutdown if we receive more than one signal.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
	<-c
	go func() {
		<-c
		os.Exit(0)
	}()

	shutdown()
}
