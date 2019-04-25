module github.com/pipelined/phono

go 1.12

require (
	github.com/go-audio/aiff v1.0.0 // indirect
	github.com/go-audio/wav v1.0.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190411002643-bd77b112433e // indirect
	github.com/gopherjs/gopherwasm v1.1.0 // indirect
	github.com/hajimehoshi/go-mp3 v0.2.0 // indirect
	github.com/hajimehoshi/oto v0.3.3 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattetti/audio v0.0.0-20190404201502-c6aebeb78429 // indirect
	github.com/pipelined/mp3 v0.0.0-20190424060305-721e3db900a9
	github.com/pipelined/pipe v0.4.0
	github.com/pipelined/signal v0.0.0-20190411172221-40f38ff7f90f
	github.com/pipelined/wav v0.0.0-20190424055427-57acedfb737a
	github.com/rs/xid v1.2.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	golang.org/x/crypto v0.0.0-20190424203555-c05e17bb3b2d // indirect
	golang.org/x/exp v0.0.0-20190424083841-8c7d1c524af6 // indirect
	golang.org/x/image v0.0.0-20190424155947-59b11bec70c7 // indirect
	golang.org/x/mobile v0.0.0-20190415191353-3e0bab5405d6 // indirect
	golang.org/x/net v0.0.0-20190424112056-4829fb13d2c6 // indirect
	golang.org/x/sys v0.0.0-20190425045458-9f0b1ff7b46a // indirect
	golang.org/x/text v0.3.1 // indirect
	golang.org/x/tools v0.0.0-20190425001055-9e44c1c40307 // indirect
)

replace (
	github.com/pipelined/mp3 => ../mp3
	github.com/pipelined/pipe => ../pipe
	github.com/pipelined/signal => ../signal
	github.com/pipelined/wav => ../wav
)
