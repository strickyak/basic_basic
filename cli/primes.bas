10 FOR i = 2 TO 1000000
20 FOR j = 2 TO i-1
30 IF i % j == 0 THEN 99
40 NEXT j
50 PRINT j
99 NEXT i
