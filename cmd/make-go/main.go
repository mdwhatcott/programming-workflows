package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mdw-tools/tcr/exec"
	"github.com/mdw-tools/tcr/gotest"
)

var Version = "dev"

const usageInfo = `
	This program runs the following go utilities in the working directory:

	- go version
	- go mod tidy
	- go fmt ./...
	- go test {args}

	The go test command {args} default to '-cover ./...' if no args are provided.
`

func main() {
	log.SetFlags(log.Ltime | log.Llongfile)
	flags := flag.NewFlagSet("make-go", flag.ContinueOnError)
	flags.Usage = func() {
		_, _ = fmt.Fprintf(
			flags.Output(),
			"Usage of make-go (version: %s)\n%s",
			Version,
			usageInfo,
		)
		flags.PrintDefaults()
	}
	err := flags.Parse(os.Args)
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(1)
	}
	args := argsForGoTest(os.Args[1:])
	path := moduleRoot()

	fmt.Println(exec.RunFatal("go version"))
	output := run(path,
		"go mod tidy",
		"go fmt ./...",
		"go test "+args,
	)
	fmt.Println("----")
	fmt.Println(strings.TrimSpace(gotest.Format(output)))
}

func moduleRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicln("os.Getwd() failed:", err)
	}
	log.Println("Looking for module root, starting at:", wd)

	for x := 0; x < 10; x++ {
		_, err := os.Stat(filepath.Join(wd, "go.mod"))
		if errors.Is(err, os.ErrNotExist) {
			wd = filepath.Dir(wd)
			continue
		}
		return wd
	}
	panic("failed to find go.mod")
}

func argsForGoTest(rawArgs []string) string {
	args := strings.Join(rawArgs, " ")
	if len(args) == 0 {
		args = "-coverprofile=/tmp/coverage.out -short -timeout=10s ./..."
	}
	return args
}

func run(working string, commands ...string) string {
	b := new(bytes.Buffer)
	for _, command := range commands {
		fmt.Println(command)
		b.WriteString(exec.RunFatal(command, exec.At(working), exec.Out(os.Stdout)))
		b.WriteString("\n")
	}
	return b.String()
}
