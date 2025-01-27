Block 0
000: de: a = m
001: b = n
002: x = 1
003: y = 0
004: r = n
005: t1 = m - 1
006: s = t1

Block 1
007: L1: if> b, 0 goto L2

Block 2
008: goto L3

Block 3
009: L2: t2 = a % b
010: reszta = t2
011: t3 = a / b
012: iloraz = t3
013: a = b
014: b = reszta
015: rr = r
016: t4 = iloraz * r
017: tmp = t4
018: if< x, tmp goto L4

Block 4
019: goto L6

Block 5
020: L4: t5 = n * iloraz
021: r = t5
022: goto L5

Block 6
023: L6: r = 0

Block 7
024: L5: t6 = r + x
025: r = t6
026: t7 = r - tmp
027: r = t7
028: ss = s
029: t8 = iloraz * s
030: tmp = t8
031: if< y, tmp goto L7

Block 8
032: goto L9

Block 9
033: L7: t9 = m * iloraz
034: s = t9
035: goto L8

Block 10
036: L9: s = 0

Block 11
037: L8: t10 = s + y
038: s = t10
039: t11 = s - tmp
040: s = t11
041: x = rr
042: y = ss
043: goto L1

Block 12
044: L3: z = a
045: ret

Block 13
046: main: read m
047: read n
048: param m
049: param n
050: param x
051: param y
052: param nwd
053: call de

Block 14
054: write x
055: write y
056: write nwd
057: halt