package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeshaw/envdecode"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"coolstercodes/modules/modulir"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Main
//
//
//
//////////////////////////////////////////////////////////////////////////////

func main() {
	rootCmd := &cobra.Command{
		Use:   "coolstercodes",
		Short: "Coolster Codes is my blog",
		Long:  "Coolster Codes is my blog, see https://coolstercodes.com",
	}

	buildCommand := &cobra.Command{
		Use:   "build",
		Short: "Run a single build loop",
		Long: strings.TrimSpace(`
Starts the build loop that watches for local changes and runs
when they're detected. A webserver is started on PORT (default
5002).`),
		Run: func(_ *cobra.Command, _ []string) {
			modulir.Build(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(buildCommand)

	loopCommand := &cobra.Command{
		Use:   "loop",
		Short: "Start build and serve loop",
		Long: strings.TrimSpace(`
Runs the build loop one time and places the result in TARGET_DIR
(default ./public/).`),
		Run: func(_ *cobra.Command, _ []string) {
			modulir.BuildLoop(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(loopCommand)

	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v", err)
		os.Exit(1)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Conf contains configuration information for the command. It's extracted from
// environment variables.
type Conf struct {
	// AbsoluteURL is the absolute URL where the compiled site will be hosted.
	// It's used for things like Atom feeds and sending email.
	AbsoluteURL string `env:"ABSOLUTE_URL,default=https://coolstercodes.com"`

	// Concurrency is the number of build Goroutines that will be used to
	// perform build work items.
	Concurrency int `env:"CONCURRENCY,default=30"`

	// Port is the port on which to serve HTTP when looping in development.
	Port int `env:"PORT,default=5002"`

	// CCEnv is the environment to run the app with. Use "development" to
	// activate development features.
	CCEnv string `env:"CC_ENV,default=development"`

	// TargetDir is the target location where the site will be built to.
	TargetDir string `env:"TARGET_DIR,default=./public"`

	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	ccEnvDevelopment = "development"
)

func getLog() *logrus.Logger {
	log := logrus.New()

	if conf.Verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}

// getModulirConfig interprets Conf to produce a configuration suitable to pass
// to a Modulir build loop.
func getModulirConfig() *modulir.Config {
	return &modulir.Config{
		Concurrency: conf.Concurrency,
		Log:         getLog(),
		LogColor:    term.IsTerminal(int(os.Stdout.Fd())),
		Port:        conf.Port,
		SourceDir:   ".",
		TargetDir:   conf.TargetDir,
		Websocket:   conf.CCEnv == ccEnvDevelopment,
	}
}
