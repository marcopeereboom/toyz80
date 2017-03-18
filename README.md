# toyz80
Emulate Z80/i8080 CPU.

This is a toy.  It is not complete so don't use it, yet!  Do send me PRs!

Currently only UT (Unit Test works) is rigged in.

1. Install Go.
2. go get github.com/marcopeereboom/toyz80
3. cd $GOPATH/src/github.com/marcopeereboom/toyz80/z80
4. go test -v

Adding more opcodes is typically a 3 step process.
1. Add opcode in z80/opcodes.go in the right location.  Array location dictates opcode!
2. Add support for opcode in z80/z80.go in the large switch statement inside func Step.
3. Add UT test case for all edge cases for new opcode in z80/z80_test.go.

If it works, send me a PR.
