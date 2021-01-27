package dhall2ds

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/briandowns/spinner"
	"github.com/inconshreveable/log15"
	gitignore "github.com/sabhiram/go-gitignore"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const ShortDescription = "exports a COMKIR Dhall record to a directory tree of YAML manifests"

var (
	destinationPath string
	timeout         time.Duration
	ignore          []string

	printHelp bool

	flagSet *flag.FlagSet
)

func logFatal(message string, ctx ...interface{}) {
	log15.Error(message, ctx...)
	os.Exit(1)
}

func usageArgs() string {
	b := bytes.Buffer{}
	w := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	fmt.Fprintln(w, "\t<path>\t(required) Dhall file to process")
	fmt.Fprintln(w, "\t<output>\t(required) destination directory")
	w.Flush()

	return fmt.Sprintf("ARGS:\n%s", b.String())
}

func Main(args []string) {
	flagSet = flag.NewFlagSet("dhall2ds", flag.ExitOnError)

	flagSet.StringVarP(&destinationPath, "output", "o", "", "(required) path to a destination directory")
	flagSet.DurationVar(&timeout, "timeout", 5*time.Minute, "length of time to run dhall command before timing out")
	flagSet.StringArrayVarP(&ignore, "ignore", "i", nil, "omit output for resources matching one of the ignore COMKIR paths. specify path with '/' separator. uses gitignore semantics for matching")
	flagSet.BoolVarP(&printHelp, "help", "h", false, "print usage instructions")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "dhall2ds %s\n", ShortDescription)
		fmt.Fprintf(os.Stderr, "Usage of ds-to-dhall dhall2ds: --output <output> <path>\n")
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flagSet.PrintDefaults()
		fmt.Fprintln(os.Stderr, usageArgs())
	}

	_ = flagSet.Parse(args)

	if printHelp {
		flagSet.Usage()
		os.Exit(0)
	}

	if destinationPath == "" {
		flagSet.Usage()
		os.Exit(1)
	}

	err := os.MkdirAll(destinationPath, 0777)
	if err != nil {
		logFatal("cannot create output directory", "err", err, "output dir", destinationPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	defer cancel()

	componentTree, err := dhallToYAML(ctx, flagSet.Arg(0))
	if err != nil {
		logFatal("failed to execute dhall-to-yaml", "error", err)
	}

	err = exportComponents(componentTree, destinationPath, ignore)

	if err != nil {
		logFatal("failed to export", "err", err)
	}
}

func dhallToYAML(ctx context.Context, dhallFile string) (map[string]interface{}, error) {
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "Running dhall-to-yaml: "
	spin.Start()
	defer spin.Stop()

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	bin := "dhall-to-yaml"
	args := []string{"--file", dhallFile}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {

		command := append([]string{bin}, args...)
		return nil, &cmdErr{
			command: strings.Join(command, " "),
			stdOut:  outBuf.String(),
			stdErr:  errBuf.String(),
		}
	}

	decoder := yaml.NewDecoder(&outBuf)

	var rv map[string]interface{}
	err = decoder.Decode(&rv)
	if err != nil {
		return nil, err
	}
	return rv, nil
}

type cmdErr struct {
	command string
	stdOut  string
	stdErr  string
}

func (c *cmdErr) Error() string {

	return strings.Join([]string{
		fmt.Sprintf("command: %q", c.command),
		fmt.Sprintf("standard out:\n%s", c.stdOut),
		fmt.Sprintf("standard err:\n%s", c.stdErr),
	}, "\n")
}

func exportYAML(contents map[string]interface{}, destinationPath string) error {
	f, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer f.Close()

	br := bufio.NewWriter(f)
	defer br.Flush()

	encoder := yaml.NewEncoder(br)
	return encoder.Encode(contents)
}

func exportComponents(componentTree map[string]interface{}, destinationPath string, ignore []string) error {
	gitIgnoreMatcher := gitignore.CompileIgnoreLines(ignore...)

	for componentName, component := range componentTree {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			return fmt.Errorf("component value for %s is not a record", componentName)
		}

		for kindName, kind := range componentMap {
			kindMap, ok := kind.(map[string]interface{})
			if !ok {
				return fmt.Errorf("kind value for %s.%s is not a record", componentName, kindName)
			}

			for resourceName, resource := range kindMap {
				if gitIgnoreMatcher.MatchesPath(filepath.Join(componentName, kindName, resourceName)) {
					continue
				}

				resourceMap, ok := resource.(map[string]interface{})
				if !ok {
					return fmt.Errorf("resource value for %s.%s.%s is not a record",
						componentName, kindName, resourceName)
				}

				dirPath := filepath.Join(destinationPath, componentName)
				err := os.MkdirAll(dirPath, 0777)
				if err != nil {
					return err
				}

				outPath := filepath.Join(dirPath, fmt.Sprintf("%s.%s.%s.yaml",
					componentName, kindName, resourceName))

				err = exportYAML(resourceMap, outPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
