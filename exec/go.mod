module github.com/voidbear-io/go/exec

go 1.18

require (
	github.com/stretchr/testify v1.11.1
	github.com/voidbear-io/go/cmdargs v0.0.0-alpha.0
	github.com/voidbear-io/go/env v0.0.0-alpha.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/voidbear-io/go/cmdargs => ../cmdargs

replace github.com/voidbear-io/go/env => ../env
