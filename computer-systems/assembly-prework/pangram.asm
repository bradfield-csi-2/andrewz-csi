section .text
global pangram
pangram:
	mov		rcx,		0
	mov		r9,			0
	mov		r10d,		0xf8000001 ;;bits 1-27 inclusive are 0. others are 1
loopstart:
	mov		cl,		[rdi]
	cmp		cl,		0x00
	je		break
	cmp		cl,		0x40 ;; if bytes below this range than skip
	jl		incloop
	and		cl,		0x1f ;; get only first 5 bits for bit shift
	mov		r9,		1		;; shift 1 by alphabet offset
	sal		r9,		cl	;; second part of above
	or		r10d,	r9d ;; set bit map place to seen
incloop:
	inc		rdi
	jmp		loopstart
break:
	mov		eax,	0		;; clear the register.
	cmp		r10d,	-1	;; check if all values seen (dependent on correct constant if r10d)
	sete	al
	ret

