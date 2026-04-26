package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const (
	environmentVariableVersionFile = "VERSION_FILE"
	environmentVariablePort        = "PORT"
)

// mainFlags are the configuration options for different environments.
type mainFlags struct {
	Port int
}

// newMainFlags creates a new, populated mainFlags structure.
// Fields are populated from command line arguments.
// If fields are not specified on the command line, environment variable values are used before defaulting to other defaults.
func newMainFlags(osArgs []string, osLookupEnvFunc func(string) (string, bool)) mainFlags {
	if len(osArgs) == 0 {
		osArgs = []string{""}
	}
	programArgs := osArgs[1:]
	var m mainFlags
	fs := m.newFlagSet(osLookupEnvFunc)
	fs.Parse(programArgs)
	return m
}

// newFlagSet creates a flagSet that populates the specified mainFlags.
func (m *mainFlags) newFlagSet(osLookupEnvFunc func(string) (string, bool)) *flag.FlagSet {
	fs := flag.NewFlagSet("main", flag.ExitOnError)
	fs.Usage = func() {
		usage(fs) // [lazy evaluation]
	}
	envValue := func(key, defaultValue string) string {
		if envValue, ok := osLookupEnvFunc(key); ok {
			return envValue
		}
		return defaultValue
	}
	envValueInt := func(key string, defaultValue int) int {
		v1 := envValue(key, "")
		v2, err := strconv.Atoi(v1)
		if err != nil {
			return defaultValue
		}
		return v2
	}
	fs.IntVar(&m.Port, "port", envValueInt(environmentVariablePort, 8080), "The port for server http requests.")
	return fs
}

// usage prints how to run the server to the flagset's output.
func usage(fs *flag.FlagSet) {
	envVars := []string{
		environmentVariableVersionFile,
		environmentVariablePort,
	}
	fmt.Fprintf(fs.Output(), "Runs the server\n")
	fmt.Fprintf(fs.Output(), "Reads environment variables when possible: [%s]\n", strings.Join(envVars, ","))
	fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
	fs.PrintDefaults()
}
