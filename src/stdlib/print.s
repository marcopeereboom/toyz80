	.include "stdlib.s"

	.org	$1000

start:
	jr	_start
text:	.db	"moo la la la", 0x0a, 0xff, 0x00
_start:
	ld	bc,text
	ld	a,(bc)
l1:
	cp	a
	jr	z,l2
	out	(COUT),a
	inc	bc
	jr	l1
l2:
	halt
