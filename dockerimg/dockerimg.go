package dockerimg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/inconshreveable/log15"
	flag "github.com/spf13/pflag"
)

const ShortDescription = "finds docker image tags in input files and adds them as Dhall records to a list"

func logFatal(message string, ctx ...interface{}) {
	log15.Error(message, ctx...)
	os.Exit(1)
}

func usageArgs() string {
	b := bytes.Buffer{}
	w := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	fmt.Fprintln(w, "\t<path>\t(required) dhall file to process")
	w.Flush()

	return fmt.Sprintf("ARGS:\n%s", b.String())
}

var (
	printHelp bool

	flagSet *flag.FlagSet
)

type ImageReference struct {
	Registry string
	Name     string
	Version  string
	Sha256   string
	Key      string
}

func (ir *ImageReference) FormatRegistry() string {
	if ir.Registry == "" {
		return "None Text"
	}

	return fmt.Sprintf("Some %q", ir.Registry)
}

// copied from https://github.com/retrohacker/parse-docker-image-name/blob/1d43ab3bde106d77374530b1d982d47375742672/index.js#L3
var (
	hasPort     = match(":[0-9]+")
	hasDot      = match("\\.")
	isLocalhost = match("^localhost(:[0-9]+)?$")
)

func domainIsNotHostName(s string) bool {
	return s != "" && !hasPort.MatchString(s) && !hasDot.MatchString(s) && !isLocalhost.MatchString(s)
}

func processReader(ir io.Reader, imgRefs *[]*ImageReference, seen map[string]struct{}) error {
	contents, err := ioutil.ReadAll(ir)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "image:")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		r, err := Parse(line)
		if err != nil {
			// silently skip over any parse errors (for instance - the line isn't something that contains a docker reference)
			continue
		}

		imgRef := &ImageReference{}
		named, ok := r.(Named)
		if !ok {
			// must have a name at least
			continue
		}

		tagged, ok := r.(Tagged)
		if !ok {
			// must have a tag at least
			continue
		}

		path := Path(named)

		imgRef.Name = path

		d := Domain(named)
		imgRef.Registry = d

		if domainIsNotHostName(d) {
			imgRef.Name = fmt.Sprintf("%s/%s", d, path)
			imgRef.Registry = ""
		}

		imgRef.Key = imgRef.Name

		imgRef.Version = tagged.Tag()

		if digested, ok := r.(Digested); ok {
			imgRef.Sha256 = strings.TrimPrefix(digested.Digest().String(), "sha256:")
		}

		if _, ok := seen[imgRef.Key]; !ok {
			*imgRefs = append(*imgRefs, imgRef)
			seen[imgRef.Key] = struct{}{}
		}

	}

	return nil
}

func processFile(path string, imgRefs *[]*ImageReference, seen map[string]struct{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	bf := bufio.NewReader(f)

	err = processReader(bf, imgRefs, seen)
	if err != nil {
		return fmt.Errorf("error on file %s: %w", path, err)
	}
	return nil
}

func processInputs(inputs []string, imgRefs *[]*ImageReference, seen map[string]struct{}) error {
	for _, input := range inputs {
		err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".dhall" {
				return processFile(path, imgRefs, seen)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

const imageRecordTemplate = `let images =
{
  {{range $index, $imgRef := .}} {{if gt $index 0}},{{end}} {{$imgRef.Key}} = {
         registry = {{$imgRef.FormatRegistry}}
         , name = "{{$imgRef.Name}}"
         , tag = "{{$imgRef.Version}}"
         , digest = Some "{{$imgRef.Sha256}}"
      }
  {{end}}
}
in images
`

var tmpl = template.Must(template.New("imageRecordDhall").Parse(imageRecordTemplate))

func Main(args []string, _ context.Context) {
	flagSet = flag.NewFlagSet("dockerimg", flag.ExitOnError)

	flagSet.BoolVarP(&printHelp, "help", "h", false, "print usage instructions")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "dockerimg %s\n", ShortDescription)
		fmt.Fprintf(os.Stderr, "Usage of ds-to-dhall dockerimg: <path>\n")
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flagSet.PrintDefaults()
		fmt.Fprintln(os.Stderr, usageArgs())
	}

	_ = flagSet.Parse(args)

	if printHelp {
		flagSet.Usage()
		os.Exit(0)
	}

	var imgRefs []*ImageReference
	seen := make(map[string]struct{})

	if len(flagSet.Args()) == 0 {
		err := processReader(os.Stdin, &imgRefs, seen)
		if err != nil {
			logFatal("failed to process from stdin", "err", err)
		}
	} else {
		err := processInputs(flagSet.Args(), &imgRefs, seen)
		if err != nil {
			logFatal("failed to process", "err", err)
		}
	}

	err := tmpl.Execute(os.Stdout, imgRefs)
	if err != nil {
		logFatal("failed to write to stdout", "err", err)
	}
}
