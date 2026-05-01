((label ff (lambda (x)
   (cond ((atom x) x)
         (t (ff (car x))))))
 '((a b) c))
