section .text
global transpose

transpose:
	;;rdi = in* matrix
	;;rsi = int* out matrix
	;;rdx = rows
	;;rcx = cols
	push rbx ;; push and save to use extra registers
	movsxd rcx, ecx ;; zero extend inputs to use 64 bit registers only, don't need but makes it easier to work with only one reg length
	movsxd rdx, edx
	mov r9, rcx ;;calculate end pointer for iterating through original array
	imul r9, rdx
	sal r9, 2
	mov r8, r9   ; calc offset to move back to beginning of next col in target array for conditional move
	neg r8
	add r8, 4
	lea rbx, [rdi + r9] ;; end pointer for iterating through original array
	sal rcx, 2 ;; width of original row
	lea r11, [rdi + rcx] ;; end pointer for first row in original array (beginning of second row)
	sal rdx, 2 ;;width of target row
.loop:
	cmp rdi, rbx ;; test if end pointer reached
	je .end

	xor rax, rax ;; zero rax reg
	cmp rdi, r11 ;;check if end of orig row/ target col reached
	cmove rax, r8 ;; set conditional point target at next col

	add rsi, rax ;; point target pointer at beginning of next call / add 0 if condition not met

	xor rax, rax ;; zero rax reg
	cmp rdi, r12 ;; check if end of orig row / target col reached
	cmove rax, rcx ;; set conditional offset to start of next row in orig array
	add r11, rax ;; add offset. or 0 if not met

	mov r10d, [rdi] ;; get int at pointer
	mov [rsi], r10d ;; store int a target
	add rsi, rdx ;; point target at next row in col
	add rdi, 4 ;; increment orig array element pointer
	jmp .loop ;;loop
.end:
	pop rbx
	ret
