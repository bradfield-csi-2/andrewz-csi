default rel

section .text
 
global volume
volume:
	mulss xmm0, xmm0 ;; square radius
	mulss xmm0, xmm1 ;; mult height
	mulss xmm0, [pithird] ;; mult pi constant
 	ret

pithird dd 1.0471975512
