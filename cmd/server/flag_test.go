package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestNewMainFlags(t *testing.T) {
	newMainFlagsTests := []struct {
		osArgs  []string
		envVars map[string]string
		want    mainFlags
	}{
		{ // all command line
			osArgs: []string{
				"ignored-binary-name",
				"-version-file=0",
				"-http-port=1",
				"-https-port=2",
				"-tls-cert-file=3",
				"-tls-key-file=4",
			},
			want: mainFlags{
				versionFile: "0",
				httpPort:    1,
				httpsPort:   2,
				tlsCertFile: "3",
				tlsKeyFile:  "4",
			},
		},
		{ // all environment variables
			envVars: map[string]string{
				"VERSION_FILE":  "0",
				"HTTP_PORT":     "1",
				"HTTPS_PORT":    "2",
				"TLS_CERT_FILE": "3",
				"TLS_KEY_FILE":  "4",
			},
			want: mainFlags{
				versionFile: "0",
				httpPort:    1,
				httpsPort:   2,
				tlsCertFile: "3",
				tlsKeyFile:  "4",
			},
		},
	}
	for i, test := range newMainFlagsTests {
		osLookupEnvFunc := func(key string) (string, bool) {
			v, ok := test.envVars[key]
			return v, ok
		}
		got := newMainFlags(test.osArgs, osLookupEnvFunc)
		if test.want != got {
			t.Errorf("test %v:\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

func TestNewMainFlagsPortOverride(t *testing.T) {
	envVars := map[string]string{
		"VERSION_FILE": "?",
		"HTTP_PORT":  "1",
		"HTTPS_PORT": "2",
		"PORT":       "3",
	}
	osLookupEnvFunc := func(key string) (string, bool) {
		v, ok := envVars[key]
		return v, ok
	}
	var osArgs []string
	want := mainFlags{
		versionFile: "?",
		httpPort:    -1,
		httpsPort:   3,
	}
	got := newMainFlags(osArgs, osLookupEnvFunc)
	if want != got {
		t.Errorf("port should override httpsPort and return -1 for http port\nwanted: %v\ngot:    %v", want, got)
	}
}

func TestUsage(t *testing.T) {
	osLookupEnvFunc := func(key string) (string, bool) {
		return "", false
	}
	var m mainFlags
	var portOverride int
	fs := m.newFlagSet(osLookupEnvFunc, &portOverride)
	var b bytes.Buffer
	fs.SetOutput(&b)
	fs.Init("", flag.ContinueOnError) // override ErrorHandling
	err := fs.Parse([]string{"-h"})
	if err != flag.ErrHelp {
		t.Errorf("wanted ErrHelp, got %v", err)
	}
	got := b.String()
	totalCommas := strings.Count(got, ",")
	b.Reset()
	fs.PrintDefaults()
	defaults := b.String()
	descriptionCommas := strings.Count(defaults, ",")
	envCommas := totalCommas - descriptionCommas
	wantEnvVarCount := envCommas + 2       // n+1 vars are joined with n commas, add an extra 1 for the PORT variable
	wantLineCount := 3 + wantEnvVarCount*2 // 3 initial lines, 2 lines per env var
	gotLineCount := strings.Count(got, "\n")
	if wantLineCount != gotLineCount {
		note := "this might be flaky, but it helps ensure that each environment variable is in the usage text"
		t.Errorf("wanted usage to have %v lines, but got %v lines. NOTE: %v, got:\n%v", wantLineCount, gotLineCount, note, got)
	}
}
