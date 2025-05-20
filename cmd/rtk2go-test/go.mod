module github.com/bramburn/gnssgo/cmd/rtk2go-test

go 1.22

require (
	github.com/bramburn/gnssgo/hardware/topgnss/top708 v0.0.0-00010101000000-000000000000
	github.com/bramburn/gnssgo/pkg/gnssgo v0.0.0-00010101000000-000000000000
	github.com/bramburn/gnssgo/pkg/ntrip v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
	go.bug.st/serial v1.6.4
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/bramburn/gnssgo/hardware/topgnss/top708 => ../../hardware/topgnss/top708
	github.com/bramburn/gnssgo/pkg/gnssgo => ../../pkg/gnssgo
	github.com/bramburn/gnssgo/pkg/ntrip => ../../pkg/ntrip
)
