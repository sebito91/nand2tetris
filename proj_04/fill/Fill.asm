// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.
// set counter to zero
@count
M=0

(LOOP)       // main loop, alternate between PAINT and DELETE
  @KBD
  D=M
  @PAINT
  D;JGT
  @DELETE
  D;JEQ

(PAINT)
  @count
  D=M
  @SCREEN
  D=D+A
  @KBD
  D=A-D
  @LOOP      // if count + SCREEN == KBD, go back to loop
  D;JLE
  
  @count
  D=M
  @SCREEN
  A=A+D
  M=-1       // set M to count + SCREEN, paint black
  @count
  M=M+1
  @LOOP
  0;JMP

(DELETE)
  @count
  D=M
  @ZERO      // if count == 0, everything deleted so go back to LOOP
  D;JEQ

  @count
  D=M
  @SCREEN
  A=A+D
  M=0        // set M to count + SCREEN, delete black
  @count
  M=M-1
  @LOOP
  0;JMP

(ZERO)
  @SCREEN
  M=0
  @LOOP
  0;JMP

(END)
  @END
  0;JMP

