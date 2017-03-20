	.include "stdlib.s"

	.org	$1000

start:
	jr	_start
text:	.db	"moo la la la", 0x0a, 0x00
_start:
	ld	bc,text
l1:
	ld	a,(bc)
	cp	a
	jr	z,l2
	out	(COUT),a
	inc	bc
	jr	l1
l2:
	halt
