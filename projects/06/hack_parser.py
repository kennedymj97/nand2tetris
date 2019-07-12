def format_line(line: str)->str:
    line = line.replace("\n", "")
    comment_index = line.find('//')
    if comment_index != -1: line = line[0:comment_index] 
    line = line.replace(' ', '')
    return line

def get_command_type(command: str) -> str:
    if command[0] == '@':
        return 'A_COMMAND'
    else:
        return 'C_COMMAND'

#dest=comp;jmp
def split_C(command: str)->(str,str,str):
    if command.count("=") > 0 and command.count(";") > 0: 
        dest, rest = command.split("=")
        comp, jump = rest.split(";")
    elif command.count(";") == 0:
        dest, comp = command.split("=")
        jump = "null"
    elif command.count("=") == 0:
        comp, jump = command.split(";")
        dest = "null"
    return (dest, comp, jump)