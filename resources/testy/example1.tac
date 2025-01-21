de:
a = m
b = n
x = 1
y = 0
r = n
t1 = m - 1
s = t1
L1:
if> b, 0 goto L2
goto L3
L2:
t2 = a % b
reszta = t2
t3 = a / b
iloraz = t3
a = b
b = reszta
rr = r
t4 = iloraz * r
tmp = t4
if< x, tmp goto L4
goto L6
L4:
t5 = n * iloraz
r = t5
goto L5
L6:
r = 0
L5:
t6 = r + x
r = t6
t7 = r - tmp
r = t7
ss = s
t8 = iloraz * s
tmp = t8
if< y, tmp goto L7
goto L9
L7:
t9 = m * iloraz
s = t9
goto L8
L9:
s = 0
L8:
t10 = s + y
s = t10
t11 = s - tmp
s = t11
x = rr
y = ss
goto L1
L3:
z = a
main:
read m
read n
param m
param n
param x
param y
param nwd
call de
write x
write y
write nwd