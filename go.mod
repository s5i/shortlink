module github.com/s5i/shortlink

go 1.24.3

require (
	github.com/boltdb/bolt v1.3.1
	github.com/s5i/goutil/authn v0.0.0-20250720174203-34b619f8ffe4
	github.com/s5i/goutil/version v0.0.0-20250720174203-34b619f8ffe4
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/ravener/discord-oauth2 v0.0.0-20230514095040-ae65713199b3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
)

replace github.com/s5i/goutil/authn v0.0.0-20250611100915-08b0ce08f828 => ../goutil/authn
