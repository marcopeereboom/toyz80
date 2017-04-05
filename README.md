# toyz80
Emulate Z80/i8080 CPU.

This is a toy.  It is not complete so don't use it, yet!  Do send me PRs!

The idea is to get to CP/M 2.2 compatibility and then build this fictional computer in actual hardware.

Currently UT (Unit Test) is rigged in and there is a basic Z80 computer being emulated that has a console and supports about 75% of the opcodes.

In order to play with this code follow the following steps:
1. Install Go.
2. `go get github.com/marcopeereboom/toyz80`

In order to test the Z80 CPU:
1. `cd $GOPATH/src/github.com/marcopeereboom/toyz80/z80`
2. `go test -v`

Adding more opcodes is typically a 3 step process.
1. Add opcode in z80/opcodes.go in the right location.  Array location dictates opcode!
2. Add support for opcode in z80/z80.go in the large switch statement inside func Step.
3. Add UT test case for all edge cases for new opcode in z80/z80_test.go.

Multi byte opcodes are a bit more involved and opcodes that don't have an example may need some extra magic.
If it works, send me a PR.

There now is a minimal implementation of a fictional computer.  In order to play with it follow these steps:
1. `cd $GOPATH/src/github.com/marcopeereboom/toyz80`
2. `go build`
3. `./toyz80 device=console,0x02-0x02 device=ram,0x0000-65536 load=0,src/cpuville/tinybasic2dms.bin`

This launches the toy z80 computer with tiny basic at address 0.  The machine will now wait for you to connect the console.  The console is a unix socket hard coded at /tmp/toyz80.socket.  Connecting to this socket can be done using socat in the following manner: "socat /dev/tty,rawer UNIX-CLIENT:/tmp/toyz80.socket".  The console is where the machine output goes.

At this point the z80 computer is ready to be either started or debugged etc.

Currently the following commands are supported in the control window:
* `mode <emacs|vi>`
* `bp [set|del address]`
* `continue`
* `disassemble [address [count]]`
* `dump [address [count]]`
* `pause`
* `registers`
* `step [count]`
* `pc <address>`
