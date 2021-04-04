section .text
global fib_recur

fib_recur:
	mov eax, edi
	cmp edi, 1
	jle .end
	push rbx
	mov ebx, edi
	lea edi, [ebx-1]
	call fib_recur
	lea edi, [ebx-2]
	mov	ebx, eax
	call fib_recur
	add eax, ebx 
	pop rbx
.end:
	ret
