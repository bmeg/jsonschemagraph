module github.com/bmeg/jsonschemagraph

go 1.20

require (
	github.com/bmeg/golib v0.0.0-20200725232156-e799a31439fc
	github.com/bmeg/grip v0.0.0-20210910231938-94d69d94ff65
	github.com/santhosh-tekuri/jsonschema/v5 v5.1.1
	github.com/spf13/cobra v1.4.0
	google.golang.org/protobuf v1.30.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/logrusorgru/aurora v0.0.0-20190428105938-cea283e61946 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	golang.org/x/crypto v0.0.0-20220507011949-2cf3adece122 // indirect
	golang.org/x/sys v0.0.0-20220908164124-27713097b956 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/santhosh-tekuri/jsonschema/v5 v5.1.1 => github.com/bmeg/jsonschema/v5 v5.3.3
