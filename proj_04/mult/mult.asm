// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Mult.asm

// Multiplies R0 and R1 and stores the result in R2.
// (R0, R1, R2 refer to RAM[0], RAM[1], and RAM[2], respectively.)

// Put your code here.
// load 0 into R2 to start
@R2
M=0

// load counter into R3
@count
M=0

// if R0 == 0, value is zero so go to end 
@R0
D=M
@END
D;JLE

// if R0 == 1, take R1 and finish
D=D-1
@LOADR1
D;JLE

// if R1 == 0, value is zero so go to end 
@R1
D=M
@END
D;JLE

// if R1 == 1, take R0 and finish
D=D-1
@LOADR0
D;JLE

// start the normal loop
@LOOP
0;JMP

(LOADR1)
@R1
D=M
@R2
M=D+M
@END
0;JMP

(LOADR0)
@R0
D=M
@R2
M=D+M
@END
0;JMP


// begin super inefficient loop
(LOOP)
  @R1
  D=M
  @count
  D=D-M     // if R1 - count <= 0, jump to end 
  @END
  D;JLE

  @R0
  D=M
  @R2
  M=D+M	    // add R0 to R2 sum, store the value back in R2
  @count
  M=M+1     // count++
  @LOOP
  0;JMP

(END) // infinite loop
  @END
  0;JMP

