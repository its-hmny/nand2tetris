// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen
// by writing 'black' in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen by writing
// 'white' in every pixel;
// the screen should remain fully clear as long as no key is pressed.


// Prelude:
// - sets the 'color' to white (the emulator's default)
// - allocates a location where the previously  pressed keycode will be saved
@color 
M = 0
@prev_key 
M = 0

(MAIN_LOOP)
    // 'input() - prev_char' will be 0 when prev_char == input()
    @KBD
    D = M
    @prev_key
    D = D - M
    // If the subtraction is 0 (prev == new), we just iterate again
    @MAIN_LOOP 
    D; JEQ
    // Updates the 'prev_key' location with the newly pressed keycode
    @KBD 
    D = M
    @prev_key
    M = D
    // Inverts the current color selection before going to the paint loop
    @color 
    M = !M
    // Then we move to the paint loop, that will set the SCREEN mmap
    @PAINT_SCREEN_LOOP 
    0; JMP


(PAINT_SCREEN_LOOP)
    // Function prelude: cleans the 'counter' and 'R0' from thei values 
    // taken from the previous computation of 'PAINT_SCREEN_LOOP'.
    @counter
    M = 0
    @R0
    M = 0
    // Iterates over every pixel in the screen memory map and sets it to the 'color'
    (LOOP)
        // Offset calculation for location 'SCREEN[i]' 
        @SCREEN 
        D = A
        @counter
        D = D + M
        // Saves location to be written at 'R0' for later usage
        @R0 
        M = D
        // Paints that specific portion of the SCREEN by setting the bytes
        @color 
        D = M
        @R0
        A = M
        M = D
        // Increments the counter and checks if another cycle is needed
        @counter
        M = M + 1
        D = M
        @8192
        D = A - D
        // Go back to the painting loop for the next portion of SCREEN
        @LOOP
        D, JGT

    // Paint loop completed, return to the caller
    @MAIN_LOOP
    0; JMP
