# toyz80
Emulate Z80/i8080 CPU.

This is a toy.  It is not complete so don't use it, yet!  Do send me PRs!

The idea is to get to CP/M 2.2 compatibility and then build this fictional computer in actual hardware.

Currently UT (Unit Test) is rigged in and there is a basic Z80 computer being emulated with minimal opcode support.

In order to play with this code follow the following steps:
1. Install Go.
2. go get github.com/marcopeereboom/toyz80

In order to test the Z80 CPU:
1. cd $GOPATH/src/github.com/marcopeereboom/toyz80/z80
2. go test -v

Adding more opcodes is typically a 3 step process.
1. Add opcode in z80/opcodes.go in the right location.  Array location dictates opcode!
2. Add support for opcode in z80/z80.go in the large switch statement inside func Step.
3. Add UT test case for all edge cases for new opcode in z80/z80_test.go.

Multi byte opcodes are a bit more involved and opcodes that don't have an example may need some extra magic.
If it works, send me a PR.

There now is a minimal implementation of a fictional computer.  In order to play with it follow these steps:
1. cd $GOPATH/src/github.com/marcopeereboom/toyz80
2. go build
3. ./toyz80
