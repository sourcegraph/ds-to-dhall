package ds2dhall

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"ds-to-dhall/comkir"
	"github.com/briandowns/spinner"
	"github.com/inconshreveable/log15"
	gitignore "github.com/sabhiram/go-gitignore"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const GeneratedComment = "{- Generated by ds-to-dhall DO NOT EDIT -}\n\n"

var (
	destinationFile string
	typeFile        string
	typesUnionFile  string
	schemaFile      string
	componentsFile  string
	timeout         time.Duration
	ignoreFiles     []string
	schemaURL       string

	printHelp bool

	flagSet *flag.FlagSet
)

const ShortDescription = "imports a deploy-sourcegraph/base into a COMKIR Dhall record"

func Main(args []string) {
	flagSet = flag.NewFlagSet("ds2dhall", flag.ExitOnError)

	flagSet.StringVarP(&destinationFile, "output", "o", "", "(required) dhall output file")
	flagSet.StringVarP(&typeFile, "type", "t", "", "dhall output type file")
	flagSet.StringVarP(&typesUnionFile, "typesUnion", "x", "", "dhall output types union file")
	flagSet.StringVarP(&schemaFile, "schema", "s", "", "dhall output schema file")
	flagSet.StringVarP(&componentsFile, "components", "c", "", "components yaml output file")
	flagSet.DurationVar(&timeout, "timeout", 5*time.Minute, "length of time to run yaml-to-dhall command before timing out")
	flagSet.StringArrayVarP(&ignoreFiles, "ignore", "i", nil, "input files matching these gitignore patterns will be ignored")
	flagSet.StringVarP(&schemaURL, "k8sSchemaURL", "u",
		"https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/a4126b7f8f0c0935e4d86f0f596176c41efbe6fe/1.18/schemas.dhall", "URL to k8s schemas.dhall file")
	flagSet.BoolVarP(&printHelp, "help", "h", false, "print usage instructions")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "ds2dhall %s\n", ShortDescription)
		fmt.Fprintf(os.Stderr, "Usage of ds-to-dhall ds2dhall: --output <output> <path>...\n")
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flagSet.PrintDefaults()
		fmt.Fprintln(os.Stderr, usageArgs())
	}

	_ = flagSet.Parse(args)

	if printHelp {
		flagSet.Usage()
		os.Exit(0)
	}

	if destinationFile == "" {
		flagSet.Usage()
		os.Exit(1)
	}

	inputs := flagSet.Args()
	if len(inputs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			logFatal("failed to get cwd for sourceDirectory", "err", err)
		}
		inputs = []string{cwd}
	}

	log15.Info("loading resources", "inputs", inputs)
	srcSet, err := loadResourceSet(inputs)
	if err != nil {
		logFatal("failed to load source resources", "error", err, "inputs", inputs)
	}

	yamlBytes, err := buildYaml(buildRecord(srcSet))
	if err != nil {
		logFatal("failed to compose yaml", "error", err)
	}

	log15.Info("execute yaml-to-dhall", "destination", destinationFile)

	dhallType := composeK8sDhallType(srcSet)
	if typeFile != "" {
		err = ioutil.WriteFile(typeFile, []byte(dhallType), 0644)
		if err != nil {
			logFatal("failed to write dhall type", "error", err, "typeFile", typeFile)
		}
		err = dhallFormat(typeFile)
		if err != nil {
			logFatal("failed to format dhall file", "error", err, "file", typeFile)
		}

		err = prependLine(typeFile, GeneratedComment)
		if err != nil {
			logFatal("failed to prepend generated comment to dhall file", "error", err, "file", typeFile)
		}
	}

	if typesUnionFile != "" {
		dhallUnionType := composeK8sDhallUnionType(srcSet)

		err = ioutil.WriteFile(typesUnionFile, []byte(dhallUnionType), 0644)
		if err != nil {
			logFatal("failed to write dhall union type", "error", err, "typesUnionFile", typesUnionFile)
		}
		err = dhallFormat(typesUnionFile)
		if err != nil {
			logFatal("failed to format dhall file", "error", err, "file", typesUnionFile)
		}

		err = prependLine(typesUnionFile, GeneratedComment)
		if err != nil {
			logFatal("failed to prepend generated comment to dhall file", "error", err, "file", typesUnionFile)
		}
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

	err = yamlToDhall(ctx, dhallType, yamlBytes, destinationFile)
	if err != nil {
		logFatal("failed to execute yaml-to-dhall", "error", err)
	}

	log15.Info("formatting output")

	err = dhallFormat(destinationFile)
	if err != nil {
		logFatal("failed to format dhall file", "error", err, "file", destinationFile)
	}

	log15.Info("prepending generated comment")

	err = prependLine(destinationFile, GeneratedComment)
	if err != nil {
		logFatal("failed to prepend generated comment to dhall file", "error", err, "file", destinationFile)
	}

	if schemaFile != "" {
		log15.Info("creating schema file")

		recordContents, err := ioutil.ReadFile(destinationFile)
		if err != nil {
			logFatal("failed to read record contents", "error", err, "destinationFile", destinationFile)
		}
		schemaContents := fmt.Sprintf("{ Type = %s, default = %s }", dhallType, string(recordContents))

		err = ioutil.WriteFile(schemaFile, []byte(schemaContents), 0644)
		if err != nil {
			logFatal("failed to write schema file", "error", err, "schemaFile", schemaFile)
		}

		err = dhallFormat(schemaFile)
		if err != nil {
			logFatal("failed to format dhall file", "error", err, "file", schemaFile)
		}

		err = prependLine(schemaFile, GeneratedComment)
		if err != nil {
			logFatal("failed to prepend generated comment to dhall file", "error", err, "file", schemaFile)
		}
	}

	if componentsFile != "" {
		log15.Info("creating components file")

		componentsBytes, err := buildYaml(buildComponents(srcSet))
		if err != nil {
			logFatal("failed to build components yaml", "error", err)
		}

		err = ioutil.WriteFile(componentsFile, componentsBytes, 0644)
		if err != nil {
			logFatal("failed to write components file", "error", err, "componentsFile", componentsFile)
		}
	}

	log15.Info("done")
}

func loadResource(rootDir string, filename string) (*comkir.Resource, error) {
	relPath, err := filepath.Rel(rootDir, filename)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	decoder := yaml.NewDecoder(br)

	var res comkir.Resource
	res.Source = filename
	err = decoder.Decode(&res.Contents)
	if err != nil {
		return nil, fmt.Errorf("failed to decode yaml file: %s: %v", filename, err)
	}

	kind, ok := res.Contents["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing a kind field", filename)
	}
	res.Kind = kind

	apiVersion, ok := res.Contents["apiVersion"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing a apiVersion field", filename)
	}
	res.ApiVersion = apiVersion

	res.DhallType = fmt.Sprintf("(%s).%s.Type", schemaURL, res.Kind)

	metadata, ok := res.Contents["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource %s is missing metadata", filename)
	}

	name, ok := metadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource %s is missing name field", filename)
	}
	res.Name = name

	labels, ok := metadata["labels"].(map[string]interface{})
	if !ok {
		// manifests without labels section exist
		labels = make(map[string]interface{})
	}

	componentLabel, ok := labels["app.kubernetes.io/component"].(string)
	if ok {
		res.Component = componentLabel
	} else {
		log15.Warn("deriving component from directory", "manifest", filename)
		res.Component = filepath.Dir(relPath)
		if res.Component == "." {
			res.Component = filepath.Base(rootDir)
		}
	}

	if res.Kind == "StatefulSet" {
		// patch statefulsets
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing spec section", filename)
		}
		volumeClaimTemplates, ok := spec["volumeClaimTemplates"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing volumeClaimTemplates section", filename)
		}
		for _, volumeClaimTemplate := range volumeClaimTemplates {
			vct, ok := volumeClaimTemplate.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("resource %s is missing volumeClaimTemplate section", filename)
			}
			vct["apiVersion"] = "v1"
			vct["kind"] = "PersistentVolumeClaim"
		}
	} else if res.Kind == "CronJob" {
		// patch cronjob
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing spec section", filename)
		}
		jobTemplateSpec, ok := spec["jobTemplate"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing jobTemplate section", filename)
		}

		_, ok = jobTemplateSpec["metadata"].(map[string]interface{})
		if !ok {
			jobTemplateSpec["metadata"] = make(map[string]interface{})
		}
	} else if res.Kind == "PersistentVolume" {
		// patch persistentvolume
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource %s is missing spec section", filename)
		}
		claimRef, ok := spec["claimRef"].(map[string]interface{})
		if ok {
			claimRef["apiVersion"] = "v1"
			claimRef["kind"] = "PersistentVolumeClaim"
		}
	}

	return &res, err
}

func usageArgs() string {
	b := bytes.Buffer{}
	w := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	fmt.Fprintln(w, "\t<path>\t(required) list of Kubernetes YAML files (or directories containing them) to process")
	fmt.Fprintln(w, "\t<output>\t(required) dhall output file")
	w.Flush()

	return fmt.Sprintf("ARGS:\n%s", b.String())
}

func makeAbs(paths []string) ([]string, error) {
	var pas []string

	for _, path := range paths {
		pa, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		pas = append(pas, pa)
	}
	return pas, nil
}

func commonPrefix(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}

	cp := strings.Split(paths[0], string(os.PathSeparator))

	if len(cp) == 0 || (len(cp) == 1 && cp[0] == "") {
		return "/", nil
	}

	for _, path := range paths[1:] {
		ps := strings.Split(path, string(os.PathSeparator))
		if len(cp) > len(ps) {
			cp = cp[:len(ps)]
		}

		idx := 0
		for idx < len(cp) && cp[idx] == ps[idx] {
			idx++
		}
		cp = cp[:idx]
	}
	if len(cp) == 0 || (len(cp) == 1 && cp[0] == "") {
		return "/", nil
	}
	return strings.Join(cp, string(os.PathSeparator)), nil
}

func loadResourceSet(inputs []string) (*comkir.ResourceSet, error) {
	pas, err := makeAbs(inputs)
	if err != nil {
		return nil, err
	}
	cr, err := commonPrefix(pas)
	if err != nil {
		return nil, err
	}
	var rs comkir.ResourceSet
	rs.Components = make(map[string][]*comkir.Resource)
	rs.Root = cr
	gitIgnoreMatcher := gitignore.CompileIgnoreLines(ignoreFiles...)

	for _, input := range pas {
		err = filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ignore := gitIgnoreMatcher.MatchesPath(path)
			if ignore && info.IsDir() {
				return filepath.SkipDir
			}
			if ignore {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
				res, err := loadResource(rs.Root, path)
				if err != nil {
					return err
				}
				rs.Components[res.Component] = append(rs.Components[res.Component], res)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return &rs, nil
}

func composeK8sDhallType(rs *comkir.ResourceSet) string {
	var schemas []string

	for component, resources := range rs.Components {
		for _, r := range resources {
			s := fmt.Sprintf("{ %s : { %s : { %s : %s } } }", component, r.Kind, r.Name, r.DhallType)
			schemas = append(schemas, s)
		}
	}

	return strings.Join(schemas, " //\\\\ ")
}

func composeK8sDhallUnionType(rs *comkir.ResourceSet) string {
	seen := make(map[string]bool)
	parts := make([]string, 0, 32)

	for _, resources := range rs.Components {
		for _, r := range resources {
			if seen[r.Kind] {
				continue
			}
			seen[r.Kind] = true
			s := fmt.Sprintf("%s : %s", r.Kind, r.DhallType)
			parts = append(parts, s)
		}
	}

	log15.Info("kubernetes union type", "size", len(seen))

	joined := strings.Join(parts, " | ")
	return "< " + joined + " >"
}

func buildRecord(rs *comkir.ResourceSet) map[string]interface{} {
	record := make(map[string]interface{})

	for component, resources := range rs.Components {
		compRec := make(map[string]map[string]interface{})
		record[component] = compRec
		for _, r := range resources {
			kindRec := compRec[r.Kind]
			if kindRec == nil {
				kindRec = make(map[string]interface{})
				compRec[r.Kind] = kindRec
			}
			kindRec[r.Name] = r.Contents
		}
	}

	return record
}

func buildYaml(record map[string]interface{}) ([]byte, error) {
	var b bytes.Buffer
	e := yaml.NewEncoder(&b)

	err := e.Encode(record)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func yamlToDhall(ctx context.Context, schema string, yamlBytes []byte, dst string) error {
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "Running yaml-to-dhall: "
	spin.Start()
	defer spin.Stop()

	var cmd *exec.Cmd
	if schema == "" {
		cmd = exec.CommandContext(ctx, "yaml-to-dhall", "--records-loose", "--output", dst)
	} else {
		// write type into a temp file because passing it on the command line exceeds the limit of characters from some shells
		typeFile, err := ioutil.TempFile("", "ds-to-dhall-type-")
		if err != nil {
			return err
		}
		defer os.Remove(typeFile.Name())

		err = ioutil.WriteFile(typeFile.Name(), []byte(schema), 0644)
		if err != nil {
			return err
		}

		cmd = exec.CommandContext(ctx, "yaml-to-dhall", typeFile.Name(), "--records-loose", "--output", dst)
	}
	cmd.Stdin = bytes.NewReader(yamlBytes)
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func dhallFormat(file string) error {
	cmd := exec.Command("dhall", "format", "--inplace", file)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func prependLine(file string, line string) error {
	tmpFile, err := ioutil.TempFile("", "ds-to-dhall-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(line)
	if err != nil {
		return err
	}

	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	cmd := exec.Command("cp", tmpFile.Name(), file)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func logFatal(message string, ctx ...interface{}) {
	log15.Error(message, ctx...)
	os.Exit(1)
}

func buildComponents(rs *comkir.ResourceSet) map[string]interface{} {
	record := make(map[string]interface{})

	for component, resources := range rs.Components {
		compRec := make(map[string]map[string]interface{})
		record[component] = compRec
		for _, r := range resources {
			kindRec := compRec[r.Kind]
			if kindRec == nil {
				kindRec = make(map[string]interface{})
				compRec[r.Kind] = kindRec
			}
			km := make(map[string]interface{})
			kindRec[r.Name] = km
			if r.Kind == "Deployment" || r.Kind == "StatefulSet" || r.Kind == "DaemonSet" {
				containers := make(map[string]interface{})
				found := extractContainersMap(r.Contents, containers)
				if found {
					km["containers"] = containers
				}
			}
		}
	}

	return record
}

func extractContainersMap(contents, containers map[string]interface{}) bool {
	for k, v := range contents {
		cm, ok := v.(map[string]interface{})

		if k == "containers" && ok {
			for ck := range cm {
				containers[ck] = struct{}{}
			}
			return true
		}

		if ok {
			found := extractContainersMap(cm, containers)
			if found {
				return true
			}
		}
	}

	return false
}
