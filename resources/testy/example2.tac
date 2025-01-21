pa:
t1 = a + b
a = t1
t2 = a - b
b = t2
pb:
param a
param b
call pa
param a
param b
call pa
pc:
param a
param b
call pb
param a
param b
call pb
param a
param b
call pb
pd:
param a
param b
call pc
param a
param b
call pc
param a
param b
call pc
param a
param b
call pc
main:
read a
read b
param a
param b
call pd
write a
write b