section .text
global fib

fib:
	mov eax, edi
	;;xor eax, eax
	mov esi, 2
	mov edx, 1
	mov ecx, 0
.test:
  cmp edi, esi
	jl .end
	lea eax, [edx + ecx]	
	mov ecx, edx
	mov edx, eax
	inc esi
	jmp .test
.end:
	ret
