; Local bindings must shadow the McCarthy primitives.
; The outer parameter `car` is bound to a wrap-in-list lambda;
; (car 'x) inside the body must dispatch to the local binding,
; not the primitive opcode.
((lambda (car) (car (quote x)))
 (lambda (z) (cons z (quote ()))))
