; A primitive name must be loadable as a value and callable indirectly.
; Both backends must define `car` (and friends) in the global env so
; that ((lambda (f) (f xs)) car) reaches the primitive via env lookup.
((lambda (f) (f (quote (a b c)))) car)
