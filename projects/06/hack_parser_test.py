from hack_parser import format_line, split_C, get_command_type
import unittest

class TestParser(unittest.TestCase):
    def test_remove_comment(self):
        self.assertEqual(format_line("// test a basic comment is removed and would be ignored"), "")
        self.assertEqual(format_line("D=M+1  // test comments after some code is removed"), "D=M+1")
        

    def test_remove_whitespace(self):
        self.assertEqual(format_line("D          =     M +         1"), "D=M+1")

    def test_ignore_whitespace(self):
        lines = ["// This line is a comment", "", "         "]
        skipped = True
        for line in lines:
            line = format_line(line)
            if len(line) == 0: continue
            skipped = False
        self.assertEqual(skipped, True)


    def test_general(self):
        self.assertEqual(format_line("@1234"), "@1234")

    def test_get_command_type(self):
        self.assertEqual(get_command_type("@1234"), "A_COMMAND")
        self.assertEqual(get_command_type("D=M+1"), "C_COMMAND")
        self.assertEqual(get_command_type("0;JMP"), "C_COMMAND")

    def test_split_C(self):
        self.assertEqual(split_C("a=b;c"), ("a", "b", "c"))
        self.assertEqual(split_C("a=b"), ("a", "b", "null"))
        self.assertEqual(split_C("b;c"), ("null", "b", "c"))

if __name__== '__main__':
    unittest.main()