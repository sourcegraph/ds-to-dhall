package dockerimg

import (
	"bufio"
	"bytes"
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

func processReader(ir io.Reader, imgRefs *[]*ImageReference, seen map[string]struct{}) error {
	contents, err := ioutil.ReadAll(ir)
	if err != nil {
		return err
	}

	matches := NotAnchoredReferenceRegexp.FindAllStringSubmatch(string(contents), -1)

	for _, match := range matches {
		if len(match) != 4 {
			continue
		}

		imgRef := &ImageReference{}

		if strings.HasPrefix(match[3], "sha256:") {
			nameParts := strings.Split(match[1], "/")
			if len(nameParts) > 1 {
				imgRef.Registry = nameParts[0]
				imgRef.Name = strings.Join(nameParts[1:], "/")
			} else {
				imgRef.Name = match[1]
			}
			imgRef.Version = match[2]
			imgRef.Sha256 = strings.TrimPrefix(match[3], "sha256:")

			if strings.HasPrefix(imgRef.Name, "sourcegraph/") {
				imgRef.Key =
					strings.Replace(strings.TrimPrefix(imgRef.Name, "sourcegraph/"), ".", "_", -1)

				if _, ok := seen[imgRef.Key]; !ok {
					*imgRefs = append(*imgRefs, imgRef)
					seen[imgRef.Key] = struct{}{}
				}
			}
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
         registry = Some "{{$imgRef.Registry}}"
         , name = "{{$imgRef.Name}}"
         , tag = Some "{{$imgRef.Version}}"
         , digest = Some "{{$imgRef.Sha256}}"
      }
  {{end}}
}
in images
`

var tmpl = template.Must(template.New("imageRecordDhall").Parse(imageRecordTemplate))

func Main(args []string) {
	flagSet = flag.NewFlagSet("dockerimg", flag.ExitOnError)

	flagSet.BoolVarP(&printHelp, "help", "h", false, "print usage instructions")

	flagSet.Usage = func() {
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
