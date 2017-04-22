# toyz80
Emulate Z80/i8080 CPU.

This is a toy.  It is not complete yet!  Do send me PRs!

The idea is to get to CP/M 2.2 compatibility and then build this fictional computer in actual hardware.

Currently UT (Unit Test) is rigged in and there is a basic Z80 computer being
emulated that has a console and supports about 98% of the opcodes.  It passes
the famed zexdoc (borrowed from
https://github.com/anotherlin/z80emu/tree/master/testfiles) tests.  Some code
compiled with the outstanding sdcc (http://sdcc.sourceforge.net/) C compiler
works as well.

Once the last few opcodes make it in and are debugged it is time to add
interrupt support and some additional devices (HDD, tape, etc)

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

Multi byte opcodes are a bit more involved and opcodes that don't have an
example may need some extra magic.  If it works, send me a PR.

There now is a minimal implementation of a fictional computer.  In order to
play with it follow these steps:
1. `cd $GOPATH/src/github.com/marcopeereboom/toyz80`
2. `go build`
3. `./toyz80 device=console,0x02-0x02 device=ram,0x0000-65536 load=0,src/cpuville/tinybasic2dms.bin`

This launches the toy z80 computer with tiny basic at address 0.  The machine
will now wait for you to connect the console.  The console is a unix socket
hard coded at /tmp/toyz80.socket.  Connecting to this socket can be done using
socat in the following manner: `socat /dev/tty,rawer UNIX-CLIENT:/tmp/toyz80.socket`.
The console is where the machine output goes.  This may seem a little hokey but
you'll thank me later (now it *is* a real serial port that can redirected etc).

Unfortunately Windows doesn't handle sockets the way UNIX does and therefore it
does not work.  At some point this will be fixed (who uses Windows anyway?).

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

### To-Do
* More instruction tests.
* Add memory fill/load instructions to control.
* Add Assembler to control.
* Add interrupt handling.
* Verify execution times.
* Cleanup, lot's of it.
* Add final missing undocumented instructions.
* Create a BIOS.
* Create floppy driver.
* Create a monitor for the Z80 machine that can be burned to ROM.
* Run CP/M.
* Translate to hardware!

### cpuville

In `src/cpuville` you'll find some publicly available goodies.  I borrowed this
from the excellent Donn Stewart cpuville pages at: http://cpuville.com/Z80.htm

You'll find Donn's ROM monitor and tinybasic which runs on his computer.  Since
toyz80 is based of that hardware it works as expected.

Thanks Donn for your super informational pages!

### sdcc

In order to play with C on your toyz80 install sdcc.  It is installed on a
host of operating systems.  I used brew on OSX.  The standard library that
comes with sdcc is pretty awesome and I was able to make images for the
emulator without modifications.  Of course if I want a more clever crt0 or
something that is another story but to just get going it is not needed.
Essentially it boils down to overriding putchar and getchar so that you can get
a console.

To play with an example do:
```
$ cd src/sdcc
$ make
sdcc -mz80 --code-loc 0x0180 --data-loc 0x1000 hello.c
makebin hello.ihx > hello.bin
$
```

Launch toyz80 (this is the control window):
```
$ toyz80 device=console,0x02-0x02 device=ram,0x0000-65536 load=0,src/sdcc/hello.bin
awaiting console connection on: /tmp/toyz80.socket
```

The emulator is now ready to be connected to.  I used `socat` for that.
```
$ socat /dev/tty,rawer UNIX-CLIENT:/tmp/toyz80.socket
```

At this point the control window is ready.  Type run and watch the z80 action!
```
> run
CPU halted
halt: $0187
af $00bb bc $000d de $000b hl $0000 ix $0000 iy $0000 pc $0187 sp $0000 f S-1H1-NC 
```

The serial window should be printing all kinds of goodies.
```
$ socat /dev/tty,rawer UNIX-CLIENT:/tmp/toyz80.socket
Hello Z80 C world! int 1 string haha hex ff

                                        *
                                           *
                                               *
                                                   *
                                                      *
                                                        *
                                                          *
                                                           *
                                                           *
                                                           *
                                                          *
                                                        *
                                                     *
                                                  *
                                              *
                                          *
                                      *
                                  *
                               *
                           *
                        *
                      *
                    *
                    *
                    *
                    *
                      *
                        *
                           *
                              *
                                  *
                                      *
```

### zexdoc results

In order to assemble zexdoc you have to install zmac.  I grabbed it from
http://48k.ca/zmac.html and compiled it myself and added it to the path.

For example:
```
$ cd z80/zex
$ zmac --zmac zexdoc.src
$ cp zout/zexdoc.cim zexdoc.com
```

The zexdoc tests are very good however they do not test negative indexed
operations.  I missed a signed two's complement displacement and the test
passed however C compiled code did not work right.  That was hilarious to
debug.

```
$ go test -v github.com/marcopeereboom/toyz80/z80 -run=TestZ  
=== RUN   TestZexDoc
Z80doc instruction exerciser
<adc,sbc> hl,<bc,de,hl,sp>....  OK
add hl,<bc,de,hl,sp>..........  OK
add ix,<bc,de,ix,sp>..........  OK
add iy,<bc,de,iy,sp>..........  OK
aluop a,nn....................  OK
aluop a,<b,c,d,e,h,l,(hl),a>..  OK
aluop a,<ixh,ixl,iyh,iyl>.....  OK
aluop a,(<ix,iy>+1)...........  OK
bit n,(<ix,iy>+1).............  OK
bit n,<b,c,d,e,h,l,(hl),a>....  OK
cpd<r>........................  OK
cpi<r>........................  OK
<daa,cpl,scf,ccf>.............  OK
<inc,dec> a...................  OK
<inc,dec> b...................  OK
<inc,dec> bc..................  OK
<inc,dec> c...................  OK
<inc,dec> d...................  OK
<inc,dec> de..................  OK
<inc,dec> e...................  OK
<inc,dec> h...................  OK
<inc,dec> hl..................  OK
<inc,dec> ix..................  OK
<inc,dec> iy..................  OK
<inc,dec> l...................  OK
<inc,dec> iy..................  OK
<inc,dec> l...................  OK
<inc,dec> (hl)................  OK
<inc,dec> sp..................  OK
<inc,dec> (<ix,iy>+1).........  OK
<inc,dec> ixh.................  OK
<inc,dec> ixl.................  OK
<inc,dec> iyh.................  OK
<inc,dec> iyl.................  OK
ld <bc,de>,(nnnn).............  OK
ld hl,(nnnn)..................  OK
ld sp,(nnnn)..................  OK
ld <ix,iy>,(nnnn).............  OK
ld (nnnn),<bc,de>.............  OK
ld (nnnn),hl..................  OK
ld (nnnn),sp..................  OK
ld (nnnn),<ix,iy>.............  OK
ld <bc,de,hl,sp>,nnnn.........  OK
ld <ix,iy>,nnnn...............  OK
ld a,<(bc),(de)>..............  OK
ld <b,c,d,e,h,l,(hl),a>,nn....  OK
ld (<ix,iy>+1),nn.............  OK
ld <b,c,d,e>,(<ix,iy>+1)......  OK
ld <h,l>,(<ix,iy>+1)..........  OK
ld a,(<ix,iy>+1)..............  OK
ld <ixh,ixl,iyh,iyl>,nn.......  OK
ld <bcdehla>,<bcdehla>........  OK
ld <bcdexya>,<bcdexya>........  OK
ld a,(nnnn) / ld (nnnn),a.....  OK
ldd<r> (1)....................  OK
ldd<r> (2)....................  OK
ldi<r> (1)....................  OK
ldi<r> (2)....................  OK
neg...........................  OK
<rrd,rld>.....................  OK
<rlca,rrca,rla,rra>...........  OK
shf/rot (<ix,iy>+1)...........  OK
shf/rot <b,c,d,e,h,l,(hl),a>..  OK
<set,res> n,<bcdehl(hl)a>.....  OK
<set,res> n,(<ix,iy>+1).......  OK
ld (<ix,iy>+1),<b,c,d,e>......  OK
ld (<ix,iy>+1),<h,l>..........  OK
ld (<ix,iy>+1),a..............  OK
ld (<bc,de>),a................  OK
Tests complete--- PASS: TestZexDoc (330.22s)
PASS
ok      github.com/marcopeereboom/toyz80/z80    330.241s
```
