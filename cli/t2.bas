10 for i = 0 to 9
20   for j = 0 to 9
30     for k = 0 to 9
40       let kk = 9 - k
44       call triangle (i*10,k+j*10,  9+i*10,j*10,  9+i*10,9+j*10, i+j*10+kk*100)
70     next k
80   next j
90 next i
