package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const (
	environmentVariableVersionFile = "VERSION_FILE"
	environmentVariableHTTPPort    = "HTTP_PORT"
	environmentVariableHTTPSPort   = "HTTPS_PORT"
	environmentVariablePort        = "PORT"
	environmentVariableTLSCertFile = "TLS_CERT_FILE"
	environmentVariableTLSKeyFile  = "TLS_KEY_FILE"
)

// mainFlags are the configuration options for different environments.
type mainFlags struct {
	versionFile string
	httpPort    int
	httpsPort   int
	tlsCertFile string
	tlsKeyFile  string
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
	portOverride := 0
	fs := m.newFlagSet(osLookupEnvFunc, &portOverride)
	fs.Parse(programArgs)
	if portOverride != 0 {
		m.httpsPort = portOverride
		m.httpPort = portOverride
	}
	return m
}

// newFlagSet creates a flagSet that populates the specified mainFlags.
func (m *mainFlags) newFlagSet(osLookupEnvFunc func(string) (string, bool), portOverride *int) *flag.FlagSet {
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
	fs.StringVar(&m.versionFile, "version-file", envValue(environmentVariableVersionFile, "version"), "A file containing the version key (the first word).  Used to bust previously cached files.  Change each time a new version of the server is run.")
	fs.IntVar(&m.httpPort, "http-port", envValueInt(environmentVariableHTTPPort, 0), "The TCP port for server http requests.  All traffic is redirected to the https port.")
	fs.IntVar(&m.httpsPort, "https-port", envValueInt(environmentVariableHTTPSPort, 0), "The TCP port for server https requests.")
	fs.IntVar(portOverride, "port", envValueInt(environmentVariablePort, 0), "The single port to run the server on.  Overrides the -https-port flag.  Causes the server to not handle http requests, ignoring -http-port.")
	fs.StringVar(&m.tlsCertFile, "tls-cert-file", envValue(environmentVariableTLSCertFile, ""), "The absolute path of the certificate file to use for TLS.")
	fs.StringVar(&m.tlsKeyFile, "tls-key-file", envValue(environmentVariableTLSKeyFile, ""), "The absolute path of the key file to use for TLS.")
	return fs
}

// usage prints how to run the server to the flagset's output.
func usage(fs *flag.FlagSet) {
	envVars := []string{
		environmentVariableVersionFile,
		environmentVariableHTTPPort,
		environmentVariableHTTPSPort,
		environmentVariableTLSCertFile,
		environmentVariableTLSKeyFile,
	}
	fmt.Fprintf(fs.Output(), "Runs the server\n")
	fmt.Fprintf(fs.Output(), "Reads environment variables when possible: [%s]\n", strings.Join(envVars, ","))
	fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
	fs.PrintDefaults()
}
