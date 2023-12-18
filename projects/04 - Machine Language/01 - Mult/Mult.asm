// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Mult.asm

// Multiplies R0 and R1 and stores the result in R2.
// Assumes that R0 >= 0, R1 >= 0, and R0 * R1 < 32768.
// (R0, R1, R2 refer to RAM[0], RAM[1], and RAM[2], respectively.)

// Prelude: cleans the R2 register to avoid memory corruption
@R2
M = 0

(LOOP)
    // If (R1 == 0) goto EXIT
    @R1
    D = M 
    @EXIT 
    D; JEQ 
    // R2 = R2 + R0
    @R0
    D = M
    @R2
    M = M + D 
    // R1 = R1 - 1
    @R1
    M = M -1
    // goto LOOP (restart iteration)
    @LOOP
    0; JMP


(EXIT) // Nothing more after this, endless cycle
    @EXIT
    0; JMP