10 for i = 0 to 9
40 call triangle (10,0,  i*10,90,  0,50, 100*i)
50 for j = 0 to 9

60 call triangle (10,0,  90,i*10,  0,50, 10*i)
70 call triangle (i*10,0,  90,i*10,  0,50, 10*i+j)

98 next j
99 next i
