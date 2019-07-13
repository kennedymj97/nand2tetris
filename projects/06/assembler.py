import argparse
import os
import hack_parser as parser
import encode_binary as encoder

# take filename as an argument
ap = argparse.ArgumentParser()
ap.add_argument('filepath', metavar='fp', type=str, help='Path to the .asm file to process.')
args = ap.parse_args()

# create an output file stream
filename = os.path.splitext(os.path.basename(args.filepath))[0]
containing_folder = os.path.dirname(args.filepath)
outfile = open(containing_folder + '/' + filename + '.hack', "w")

# intitialise symbol table
symbols = {
    "SP": 0,
    "LCL": 1, 
    "ARG": 2,
    "THIS": 3,
    "THAT": 4,
    "R0": 0,
    "R1": 1,
    "R2": 2,
    "R3": 3,
    "R4": 4,
    "R5": 5,
    "R6": 6,
    "R7": 7,
    "R8": 8,
    "R9": 9,
    "R10": 10,
    "R11": 11,
    "R12": 12,
    "R13": 13,
    "R14": 14,
    "R15": 15,
    "SCREEN": 16384,
    "KBD": 24576
}

rom_address = 0
commands = []

# create an input file stream and loop through it line by line
with open(args.filepath, 'r') as infile:
    for line in infile:
        # parse each line
        command = parser.format_line(line)

        # ignore line if there is no command 
        if len(command) == 0: continue

        # check the command type
        command_type = parser.get_command_type(command)

        # produce binary output
        # write binary output to file
        if command_type == "A_COMMAND":
            commands.append(command)
            rom_address += 1
        elif command_type == "C_COMMAND":
            commands.append(command)
            rom_address += 1
        elif command_type == "L_COMMAND":
            symbol = command[1:-1]
            symbols[symbol] = rom_address
            

next_available_address = 16
for command in commands:
    # check the command type
    command_type = parser.get_command_type(command)

    # produce binary output
    # write binary output to file
    if command_type == "A_COMMAND":
        value = command[1:]
        if parser.A_is_symbol(value):
            if value not in symbols:
                symbols[value] = next_available_address
                next_available_address += 1
            address = symbols[value]
        else:
            address = int(value)
        encoded_command = encoder.encode_A(address)
        outfile.writelines(encoded_command + "\n")
    elif command_type == "C_COMMAND":
        dest, comp, jump = parser.split_C(command)
        encoded_command = encoder.encode_C(dest, comp, jump)
        outfile.writelines(encoded_command + "\n")
    
