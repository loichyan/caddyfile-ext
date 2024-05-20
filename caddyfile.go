package caddyfile_ext

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
	for i := 1; i <= 16; i++ {
		httpcaddyfile.RegisterGlobalOption("app"+strconv.Itoa(i), parseApp)
	}
}

// parseApp sets up a third-party app from the Caddyfile syntax. Syntax:
//
//	appN <name> {
//		number 123
//		boolean true
//		string abc
//		object {
//			...
//		}
//		+array 1
//	}
func parseApp(d *caddyfile.Dispenser, prev any) (any, error) {
	d.Next()
	id := d.Val()
	if prev != nil {
		return nil, d.Errf("duplicate App '%s'", id)
	}
	if !d.NextArg() {
		return nil, d.Err("missing App name")
	}
	name := d.Val()
	if !d.Next() {
		return nil, d.Err("missing App configuration")
	}
	cfg, err := parseArgs(d, "", nil)
	if err != nil {
		return nil, err
	}
	return httpcaddyfile.App{
		Name:  name,
		Value: caddyconfig.JSON(cfg, nil),
	}, nil
}

func parseArgs(d *caddyfile.Dispenser, key string, prev any) (any, error) {
	curr := d.Val()
	switch {
	case d.ValRaw() == "{":
		// step into nested object
		prev = map[string]any{}
		for {
			d.Next()
			if d.ValRaw() == "}" {
				break
			}
			updated, err := parseArgs(d, "", prev)
			if err != nil {
				return nil, err
			}
			prev = updated
		}
		if nextOnSameLine(d) {
			return nil, d.Err("unexpected value after an object")
		} else {
			return prev, nil
		}
	case nextOnSameLine(d):
		// make sure previous value is an object
		key = curr
		obj, ok := prev.(map[string]any)
		if ok {
		} else if prev == nil {
			obj = map[string]any{}
		} else {
			return nil, d.Errf("'%s' is not an object", key)
		}
		var updated any
		if key[0] == '+' {
			// update an existing array
			// ..path +key ..rest
			//         ^^^
			key = key[1:]
			prev = obj[key]
			// make sure existing value is an array
			arr, ok := prev.([]any)
			if ok {
			} else if prev == nil {
				arr = []any{}
			} else {
				return nil, d.Errf("'%s' is not an array", key)
			}
			val, err := parseArgs(d, key, nil)
			if err != nil {
				return nil, err
			}
			updated = append(arr, val)
		} else {
			// update an existing object
			// ..path key ..rest
			//        ^^^
			val, err := parseArgs(d, key, obj[key])
			if err != nil {
				return nil, err
			}
			updated = val
		}
		obj[key] = updated
		return obj, nil
	default:
		// return the parsed value
		// ..path val
		//        ^^^
		if prev != nil {
			return nil, d.Errf("duplicate value for '%s'", key)
		}
		return d.ScalarVal(), nil
	}
}

func nextOnSameLine(d *caddyfile.Dispenser) bool {
	line := d.Line()
	return d.Next() && (d.Line() == line || !d.Prev())
}
