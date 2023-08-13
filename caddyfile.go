package caddyfile_ext

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("app_ext", parseApp)
}

// parseApp sets up a third-party app from the Caddyfile syntax. Syntax:
//
//	app_ext <name> {
//		number 123
//		boolean true
//		string "..."
//		array [ 1 2 3 ]
//		object {
//		}
//	}
func parseApp(d *caddyfile.Dispenser, _ any) (any, error) {
	if !(d.Next() && d.NextArg()) {
		return nil, d.Err("missing App name")
	}
	name := d.Val()
	if !d.Next() {
		return nil, d.Err("missing App configuration")
	}
	cfg, err := parseArgs(d, nil)
	if err != nil {
		return nil, err
	}
	return httpcaddyfile.App{
		Name:  name,
		Value: caddyconfig.JSON(cfg, nil),
	}, nil
}

func parseArgs(d *caddyfile.Dispenser, prev any) (any, error) {
	// TODO: parse object, array

	curr := d.Val()
	if d.Next() {
		// create or update an existing object
		// ..path key ..val
		//        ^^^
		var obj map[string]any = nil
		if prev == nil {
			// create a new object
			obj = map[string]any{}
		} else {
			// update previous object
			prev, ok := prev.(map[string]any)
			if !ok {
				return nil, d.Err("attempt to update a non-object")
			}
			obj = prev
		}
		val, err := parseArgs(d, obj[curr])
		if err != nil {
			return nil, err
		}
		obj[curr] = val
		return obj, nil
	} else {
		// return the parsed value
		// ..path val
		//        ^^^
		if prev != nil {
			return nil, d.Err("duplicate value")
		}
		return d.ScalarVal(), nil
	}
}
