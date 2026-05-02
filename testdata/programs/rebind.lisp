; Top-level rebinding of a primitive must take effect for subsequent
; calls. After (label car ...) the (car 'a) call must dispatch to the
; rebound user lambda, not the primitive opcode.
(label car (lambda (x) (cons x (quote ()))))
(car (quote a))
