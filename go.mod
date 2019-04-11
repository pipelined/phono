module github.com/pipelined/phono

go 1.12

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pipelined/mp3 v0.0.0-20190328061808-a2441113348b
	github.com/pipelined/pipe v0.4.0
	github.com/pipelined/signal v0.0.0-20190303105250-40bacde8022c
	github.com/pipelined/wav v0.0.0-20190325201312-7b36fae928f4
	github.com/rs/xid v1.2.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
)

replace (
	github.com/pipelined/mp3 => ../mp3
	github.com/pipelined/pipe => ../pipe
	github.com/pipelined/signal => ../signal
	github.com/pipelined/wav => ../wav
)
