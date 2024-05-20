# 🍭 caddyfile-ext

Extends your Caddyfile with additional JSON configuration.

## ✍️ Example

```caddy
{
	# set up a third-party app, it supports up to 16 different apps,
	# i.e. app1, app2, ... app16
	app1 example {
		# set value of a field
		# > "field": "whatever"
		field whatever

		# value is first converted to a scalar type
		# > "boolean": true
		# > "number": 123.45
		boolean true
		number 123.45

		# quoted value is always a string
		# > "string": "false"
		string "false"

		# multiple arguments are considered as a path to update a value
		# > "a": {
		# >		"b": {
		# >			"c": {
		# >				"k1": "v1",
		# >				"k2": "v2"
		# >			}
		# >		}
		# > }
		a b c {
			k1 v2
			k2 v1
		}

		# key starts with '+' represents an array
		# > "arr": [1, 2]
		+arr 1
		+arr 2

		# '+' is available in a path to create a single-element array
		# > "obj": {
		# >		"arr": [{
		# >			"key": "val"
		# >		}]
		# > }
		obj +arr key val
	}

	# create a layer4 App
	app2 layer4 servers main {
		+listen :7443
		+routes {
			+match tls +sni xx.example.com
			+handle {
				handler proxy
				+upstreams +dial :7001
			}
		}
		+routes {
			+match tls +sni yy.example.com
			+handle {
				handler proxy
				+upstreams +dial :7002
			}
		}
		+routes {
			+handle {
				handler proxy
				+upstreams +dial :7003
			}
		}
	}
}
```

## ⚖️ License

Licensed under either of

- Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE) or
  <http://www.apache.org/licenses/LICENSE-2.0>)
- MIT license ([LICENSE-MIT](LICENSE-MIT) or <http://opensource.org/licenses/MIT>)

at your option.
