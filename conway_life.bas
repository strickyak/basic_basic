###### ###### ###### ###### ###### ######
90 DIM a(22,22)
91 DIM b(22,22)
100 LET n = 21

let a(15,10) = 1
let a(16,9) = 1
let a(16,10) = 1
let a(16,11) = 1
let a(17,11) = 1

for q = 0 to 5
let q15 = q*15
for p = 0 to 5
let p15 = p*15
gosub 2000
gosub 1200
gosub 1100
next p
next q
goto 9999

1100 REM copy b() to a()
FOR i=0 to n+1
FOR j=0 to n+1
LET a(i,j) = b(i,j)
NEXT j
NEXT i
return

1200 REM life step a() to b()
1210 FOR i=1 to n
1220 FOR j=1 to n
1230 LET k= a(i-1,j-1) + a(i-1,j) + a(i-1,j+1) + a(i,j-1) + a(i,j+1) + a(i+1,j-1) + a(i+1,j) + a(i+1,j+1) 
let k = a(i,j)*10 + k

1240 LET b(i,j) = 1
1260 IF k==3 then 1600
1261 IF k==12 then 1600
1262 IF k==13 then 1600

1270 LET b(i,j) = 0

1600 REM continue here
1610 NEXT j
1620 NEXT i
return


2000 REM display
let h = 0.5
2010 FOR i = 1 to n
let i2 = i*0.7
2020 FOR j = 1 to n
let j2 = j*0.7

let c = 393 ;REM green
if a(i,j) then 2150
let c = 222 ;REM gray
2150 call triangle (p15+i2,95-(j2+q15), p15+i2+h,95-(j2+q15), p15+i2,95-(j2+h+q15), c)

NEXT j
NEXT i
return


###### ###### ###### ###### ###### ######
90 DIM a(22,22)
91 DIM b(22,22)
100 LET n = 21

let a(15,10) = 1
let a(16,9) = 1
let a(16,10) = 1
let a(16,11) = 1
let a(17,11) = 1

for q = 0 to 5
let q15 = q*15
for p = 0 to 5
let p15 = p*15
gosub 2000
gosub 1200
gosub 1100
next p
next q
goto 9999

1100 REM copy b() to a()
FOR i=0 to n+1
FOR j=0 to n+1
LET a(i,j) = b(i,j)
NEXT j
NEXT i
return

1200 REM life step a() to b()
1210 FOR i=1 to n
1220 FOR j=1 to n
1230 LET k= a(i-1,j-1) + a(i-1,j) + a(i-1,j+1) + a(i,j-1) + a(i,j+1) + a(i+1,j-1) + a(i+1,j) + a(i+1,j+1) 
let k = a(i,j)*10 + k

1240 LET b(i,j) = 1
1260 IF k==3 then 1600
1261 IF k==12 then 1600
1262 IF k==13 then 1600

1270 LET b(i,j) = 0

1600 REM continue here
1610 NEXT j
1620 NEXT i
return


2000 REM display
let h = 0.5
2010 FOR i = 1 to n
let i2 = i*0.7
2020 FOR j = 1 to n
let j2 = j*0.7

let c = 393 ;REM green
if a(i,j) then 2150
let c = 222 ;REM gray
2150 call triangle (p15+i2,95-(j2+q15), p15+i2+h,95-(j2+q15), p15+i2,95-(j2+h+q15), c)

NEXT j
NEXT i
return


