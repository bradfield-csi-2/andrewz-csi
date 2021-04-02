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
	lea r9, [rdi + 4*rcx]
	mov rax, [r9]
	ret
