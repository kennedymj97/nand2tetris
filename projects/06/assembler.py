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
            encoded_command = encoder.encode_A(command)
            outfile.writelines(encoded_command + "\n")
        elif command_type == "C_COMMAND":
            dest, comp, jump = parser.split_C(command)
            encoded_command = encoder.encode_C(dest, comp, jump)
            outfile.writelines(encoded_command + "\n")
        
