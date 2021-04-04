section .text
global index
index:
	; rdi: matrix
	; rsi: rows
	; rdx: cols
	; rcx: rindex
	; r8: cindex
	imul ecx, edx
	add ecx, r8d
	mov rax, [rdi + 4*rcx]
	ret
