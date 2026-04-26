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
				"-port=1",
			},
			want: mainFlags{
				versionFile: "0",
				Port:        1,
			},
		},
		{ // all environment variables
			envVars: map[string]string{
				"VERSION_FILE": "0",
				"PORT":         "1",
			},
			want: mainFlags{
				versionFile: "0",
				Port:        1,
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

func TestUsage(t *testing.T) {
	osLookupEnvFunc := func(key string) (string, bool) {
		return "", false
	}
	var m mainFlags
	fs := m.newFlagSet(osLookupEnvFunc)
	var b bytes.Buffer
	fs.SetOutput(&b)
	fs.Init("", flag.ContinueOnError) // override ErrorHandling
	err := fs.Parse([]string{"-h"})
	if err != flag.ErrHelp {
		t.Errorf("wanted ErrHelp, got %v", err)
	}
	got := b.String()
	b.Reset()
	fs.PrintDefaults()
	wantEnvVarCount := 2
	wantLineCount := 3 + wantEnvVarCount*2 // 3 initial lines, 2 lines per env var
	gotLineCount := strings.Count(got, "\n")
	if wantLineCount != gotLineCount {
		note := "this might be flaky, but it helps ensure that each environment variable is in the usage text"
		t.Errorf("wanted usage to have %v lines, but got %v lines. NOTE: %v, got:\n%v", wantLineCount, gotLineCount, note, got)
	}
}
