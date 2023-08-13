package caddyfile_ext

import (
	"encoding/json"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func testParser(t *testing.T, input string, expected string) {
	d := caddyfile.NewTestDispenser(input)
	d.Next()
	out, err := parseArgs(d, nil)
	if err != nil {
		t.Error(err)
	} else {
		out, _ := json.Marshal(out)
		if string(out) != expected {
			t.Errorf("assertion failed:\nexpected: %s\n     got: %s\n", expected, out)
		}
	}
}

func TestScalarVal(t *testing.T) {
	testParser(t, `true`, `true`)
	testParser(t, `123`, `123`)
	testParser(t, `-4.5`, `-4.5`)
	testParser(t, `"false"`, `"false"`)
}

func TestPathUpdate(t *testing.T) {
	testParser(t, `key1 key2 key3 true`, `{"key1":{"key2":{"key3":true}}}`)
}
