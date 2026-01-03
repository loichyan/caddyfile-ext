package caddyfile_ext

import (
	"encoding/json"
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
		return nil, d.Errf("duplicate app `%s`", id)
	}
	p := parser{d}
	if !p.NextArg() {
		return nil, d.Err("app name is absent")
	}
	k, err := p.parseKey()
	if err != nil {
		return nil, err
	}
	cfg, err := p.parse(k, nil)
	if err != nil {
		return nil, err
	}
	return httpcaddyfile.App{
		Name:  k.val,
		Value: caddyconfig.JSON(cfg, nil),
	}, nil
}

type parser struct {
	*caddyfile.Dispenser
}

type key struct {
	plus   int
	equals bool
	val    string
}

func (p *parser) parse(k key, prev any) (any, error) {
	var curr any

	switch {
	case k.equals:
		// provide value in raw JSON
		// ..path =key <json>
		//             ^^^^^^

		err := json.Unmarshal([]byte(p.Val()), &curr)
		if err != nil {
			return nil, err
		}
		if p.hasNextCurrentLine() {
			return nil, p.Errf("unexpected arguments after JSON for `%s`", k.val)
		}

	case p.ValRaw() == "{":
		// provide value within braces
		// ..path key {...}
		//            ^^^^^

		obj := map[string]any{}
		for p.Next() && p.ValRaw() != "}" {
			err := p.parseEntry(obj)
			if err != nil {
				return nil, err
			}
		}
		if p.hasNextCurrentLine() {
			return nil, p.Errf("unexpected arguments after braces for `%s`", k.val)
		}
		curr = obj

	case p.hasNextCurrentLine():
		// update object with the specified path
		// ..path key key2 ..rest
		//            ^^^^

		if k.plus == 0 {
			obj, ok := prev.(map[string]any)
			if prev == nil {
				obj = map[string]any{}
			} else if !ok {
				return nil, p.Errf("value for `%s` is not an object", k.val)
			}
			return obj, p.parseEntry(obj)
		} else {
			obj := map[string]any{}
			k2, err := p.parseKey()
			if err != nil {
				return nil, err
			}

			obj[k2.val], err = p.parse(k2, nil)
			if err != nil {
				return nil, err
			}
			curr = obj
		}

	default:
		// provide a scalar value
		// ..path key val
		//            ^^^
		curr = p.ScalarVal()
	}

	if k.plus > 0 {
		// make sure target value is an array
		arr, ok := prev.([]any)
		if prev == nil {
			arr = []any{}
		} else if !ok {
			return nil, p.Errf("value for '%s' is not an array", k.val)
		}

		// handle nested array
		for i := k.plus; i > 1; i-- {
			curr = []any{curr}
		}
		return append(arr, curr), nil
	} else if prev != nil {
		return nil, p.Errf("duplicate value for key `%s`", k.val)
	} else {
		return curr, nil
	}
}

func (p *parser) parseEntry(obj map[string]any) error {
	k, err := p.parseKey()
	if err != nil {
		return err
	}
	val, err := p.parse(k, obj[k.val])
	if err != nil {
		return err
	}
	obj[k.val] = val
	return nil
}

func (p *parser) parseKey() (key, error) {
	k := key{}
	tok := p.Val()
	for i := 0; ; i++ {
		if tok[i] == '+' {
			k.plus += 1
		} else if tok[i] == '=' {
			k.equals = tok[i] == '='
			k.val = tok[i+1:]
			break
		} else {
			k.val = tok[i:]
			break
		}
	}
	if !p.nextCurrentLine() {
		return k, p.Errf("value for `%s` is missing", k.val)
	}
	return k, nil
}

func (p *parser) hasNextCurrentLine() bool {
	return p.nextCurrentLine() && p.Prev()
}

func (p *parser) nextCurrentLine() bool {
	line := p.Line()
	return p.Next() && (p.Line() == line || !p.Prev())
}
