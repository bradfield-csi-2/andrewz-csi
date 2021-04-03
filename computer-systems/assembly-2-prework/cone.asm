default rel

section .text
 
global volume
volume:
	mulss xmm0, xmm0 ;; square radius
	mulss xmm0, xmm1 ;; mult height
	vmovss xmm2, [pithird] ;; load pi constant
	mulss xmm0, xmm2 ;; mult pi constant
	;;vmovss xmm2, [third]
;;	mulss xmm0, xmm2
	;;vmovss xmm2, [three] ;; load 3.0
	;;divss xmm0, [three] ;;div by 3.0
 	ret

;pi dd 3.141592653589793238462 
;third dd 0.33333333
;three dd 3.0
pithird dd 1.0471975512
