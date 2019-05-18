module github.com/pipelined/phono

go 1.12

require (
	github.com/go-audio/wav v1.0.0 // indirect
	github.com/hajimehoshi/go-mp3 v0.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pipelined/mp3 v0.0.0-20190424060305-721e3db900a9
	github.com/pipelined/pipe v0.4.0
	github.com/pipelined/signal v0.0.0-20190411172221-40f38ff7f90f
	github.com/pipelined/wav v0.0.0-20190424055427-57acedfb737a
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
)

replace (
	github.com/pipelined/mp3 => ../mp3
	github.com/pipelined/pipe => ../pipe
	github.com/pipelined/signal => ../signal
	github.com/pipelined/wav => ../wav
)
