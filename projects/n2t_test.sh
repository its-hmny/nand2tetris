#! /bin/bash

set -eu

# ANSI escape codes for colors and attributes
bold=$(tput bold)
red=$(tput setaf 1)
blue=$(tput setaf 4)
green=$(tput setaf 2)
yellow=$(tput setaf 3)
reset=$(tput sgr0)

# Define blacklist of files to skip
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
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
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
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
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
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
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
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
            continue
        fi
        
        # Check if file is in the blacklist
        if [[ " ${blacklist[@]} " =~ " $tst_name " ]]; then
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
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
            echo "${blue}${bold}- Skipping test: '$tst_name'...${reset}"
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