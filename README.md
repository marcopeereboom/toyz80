# toyz80
Emulate Z80/i8080 CPU.

This is a toy.  It is not complete yet!  Do send me PRs!

The idea is to get to CP/M 2.2 compatibility and then build this fictional computer in actual hardware.

Currently UT (Unit Test) is rigged in and there is a basic Z80 computer being emulated that has a console and supports about 95% of the opcodes.  It almost passes the famed zexdoc (borrowed from https://github.com/anotherlin/z80emu/tree/master/testfiles) tests.  Some code compiled with the outstanding sdcc (http://sdcc.sourceforge.net/) C compiler works as well.  A few more opcodes need debugging and `printf("%s\r\n", "moo");` will work!

Once the last few opcodes make it in and are debugged it is time to add interrupt support and some additional devices (HDD, tape, etc)

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

This launches the toy z80 computer with tiny basic at address 0.  The machine will now wait for you to connect the console.  The console is a unix socket hard coded at /tmp/toyz80.socket.  Connecting to this socket can be done using socat in the following manner: `socat /dev/tty,rawer UNIX-CLIENT:/tmp/toyz80.socket`.  The console is where the machine output goes.  This may seem a little hokey but you'll thank me later (now it *is* a real serial port that can redirected etc).

Unfortunately Windows doesn't handle sockets the way UNIX does and therefore it does not work.  At some point this will be fixed (who uses Windows anyway?).

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

### Current zexdoc results
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
<set,res> n,(<ix,iy>+1).......  SKIPPED
ld (<ix,iy>+1),<b,c,d,e>......  OK
ld (<ix,iy>+1),<h,l>..........  OK
ld (<ix,iy>+1),a..............  OK
ld (<bc,de>),a................  OK
```
