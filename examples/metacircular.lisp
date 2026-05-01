; McCarthy's eval, written in itself.
; Built on the five primitives (car, cdr, cons, atom, eq) plus
; quote, cond, lambda, and label.

(label null  (lambda (x) (eq x '())))

(label and   (lambda (x y)
               (cond (x (cond (y 't) ('t '())))
                     ('t '()))))

(label not   (lambda (x) (cond (x '()) ('t 't))))

(label append (lambda (x y)
                (cond ((null x) y)
                      ('t (cons (car x) (append (cdr x) y))))))

(label caar   (lambda (x) (car (car x))))
(label cadr   (lambda (x) (car (cdr x))))
(label cadar  (lambda (x) (car (cdr (car x)))))
(label caddr  (lambda (x) (car (cdr (cdr x)))))
(label caddar (lambda (x) (car (cdr (cdr (car x))))))

(label pair  (lambda (x y)
               (cond ((and (null x) (null y)) '())
                     ((and (not (atom x)) (not (atom y)))
                      (cons (cons (car x) (cons (car y) '()))
                            (pair (cdr x) (cdr y)))))))

(label assoc (lambda (x y)
               (cond ((eq (caar y) x) (cadar y))
                     ('t (assoc x (cdr y))))))

(label evcon (lambda (c a)
               (cond ((eval (caar c) a) (eval (cadar c) a))
                     ('t (evcon (cdr c) a)))))

(label evlis (lambda (m a)
               (cond ((null m) '())
                     ('t (cons (eval (car m) a) (evlis (cdr m) a))))))

(label eval (lambda (e a)
              (cond
                ((atom e) (assoc e a))
                ((atom (car e))
                 (cond
                   ((eq (car e) 'quote) (cadr e))
                   ((eq (car e) 'atom)  (atom (eval (cadr e) a)))
                   ((eq (car e) 'eq)    (eq   (eval (cadr e) a)
                                              (eval (caddr e) a)))
                   ((eq (car e) 'car)   (car  (eval (cadr e) a)))
                   ((eq (car e) 'cdr)   (cdr  (eval (cadr e) a)))
                   ((eq (car e) 'cons)  (cons (eval (cadr e) a)
                                              (eval (caddr e) a)))
                   ((eq (car e) 'cond)  (evcon (cdr e) a))
                   ('t (eval (cons (assoc (car e) a) (cdr e)) a))))
                ((eq (caar e) 'label)
                 (eval (cons (caddar e) (cdr e))
                       (cons (cons (cadar e) (cons (car e) '())) a)))
                ((eq (caar e) 'lambda)
                 (eval (caddar e)
                       (append (pair (cadar e) (evlis (cdr e) a)) a))))))

(eval '(car (quote (a b c))) '())
