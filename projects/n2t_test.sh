#! /bin/bash

set -eu

# Define ANSI escape codes for colors and attributes
bold="\033[1m"
red="\033[31m"
blue="\033[34m"
green="\033[32m"
yellow="\033[33m"
reset="\033[0m"

# Define blacklist of test files to skip (they need user interaction)
# - Fill: requires user evaluation (press a key and see the screen change color)
# - Memory: requires the user to press a key in order to continue the testing
blacklist=("Fill.tst" "Memory.tst")

# Get the directory of the script
SCRIPT_DIR=$(dirname "$(realpath "$0")")


# Function to test '01 - Logic Gates' folder
test_logic_gates() {
    echo -e "${blue}${bold}Testing '01 - Logic Gates'...${reset}"

    # Change to the directory of the script
    cd "$SCRIPT_DIR" || exit

    # Iterate over each .tst file and run HardwareSimulator.sh
    for tst_path in ./01\ -\ Logic\ Gates/**/*.tst; do
        tst_name=$(basename "$tst_path")

        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo -e "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        echo -e "${bold}${yellow}- Running test: '$tst_name'...${reset}"
        ../tools/HardwareSimulator.sh "$tst_path" > /dev/null

        if [ $? -eq 0 ]; then
            echo -e "${green}${bold}  Success: '$tst_name' completed${reset}"
        else
            echo -e "${red}${bold}  Error: '$tst_name' failed${reset}"
        fi
    done
}

# Function to test '02 - Boolean Arithmetic' folder
test_boolean_arithmetic() {
    echo -e "${blue}${bold}Testing '02 - Boolean Arithmetic'...${reset}"

    # Change to the directory of the script
    cd "$SCRIPT_DIR" || exit

    # Iterate over each .tst file and run HardwareSimulator.sh
    for tst_path in ./02\ -\ Boolean\ Arithmetic/**/*.tst; do
        tst_name=$(basename "$tst_path")

        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo -e "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        echo -e "${bold}${yellow}- Running test: '$tst_name'...${reset}"
        ../tools/HardwareSimulator.sh "$tst_path" > /dev/null

        if [ $? -eq 0 ]; then
            echo -e "${green}${bold}  Success: '$tst_name' completed${reset}"
        else
            echo -e "${red}${bold}  Error: '$tst_name' failed${reset}"
        fi
    done
}

# Function to test '03 - Sequential Logic' folder
test_sequential_logic() {
    echo -e "${blue}${bold}Testing '03 - Sequential Logic'...${reset}"

    # Change to the directory of the script
    cd "$SCRIPT_DIR" || exit

    # Iterate over each .tst file and run HardwareSimulator.sh
    for tst_path in ./03\ -\ Sequential\ Logic/**/*.tst; do
        tst_name=$(basename "$tst_path")

        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo -e "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        echo -e "${bold}${yellow}- Running test: '$tst_name'...${reset}"
        ../tools/HardwareSimulator.sh "$tst_path" > /dev/null

        if [ $? -eq 0 ]; then
            echo -e "${green}${bold}  Success: '$tst_name' completed${reset}"
        else
            echo -e "${red}${bold}  Error: '$tst_name' failed${reset}"
        fi
    done
}

# Function to test '04 - Machine Language' folder
test_machine_language() {
    echo -e "${blue}${bold}Testing '04 - Machine Language'...${reset}"

    # Change to the directory of the script
    cd "$SCRIPT_DIR" || exit

    # Iterate over each .tst file and run HardwareSimulator.sh
    for tst_path in ./04\ -\ Machine\ Language/**/*.tst; do
        tst_name=$(basename "$tst_path")
        
        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo -e "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        echo -e "${bold}${yellow}- Running test: '$tst_name'...${reset}"
        ../tools/CPUEmulator.sh "$tst_path" > /dev/null

        if [ $? -eq 0 ]; then
            echo -e "${green}${bold}  Success: '$tst_name' completed${reset}"
        else
            echo -e "${red}${bold}  Error: '$tst_name' failed${reset}"
        fi
    done
}

# Function to test '05 - Computer Architecture' folder
test_computer_architecture() {
    echo -e "${blue}${bold}Testing '05 - Computer Architecture'...${reset}"

    # Change to the directory of the script
    cd "$SCRIPT_DIR" || exit

    # Rename the file
    for dep_hdl in ./05\ -\ Computer\ Architecture/**/*.Copy.hdl; do
        mv "$dep_hdl" "${dep_hdl%.Copy.hdl}.hdl"
    done

    # Iterate over each .tst file and run HardwareSimulator.sh
    for tst_path in ./05\ -\ Computer\ Architecture/**/*.tst; do
        tst_name=$(basename "$tst_path")

        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo -e "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        echo -e "${bold}${yellow}- Running test: '$tst_name'...${reset}"
        ../tools/HardwareSimulator.sh "$tst_path" > /dev/null

        if [ $? -eq 0 ]; then
            echo -e "${green}${bold}  Success: '$tst_name' completed${reset}"
        else
            echo -e "${red}${bold}  Error: '$tst_name' failed${reset}"
        fi
    done
}

functions=("test_logic_gates" "test_boolean_arithmetic" "test_sequential_logic" "test_machine_language" "test_computer_architecture")

# Iterate over the array and call each function
for func in "${functions[@]}"; do
    $func  # Calls the test function
    printf "\n\n"
done