from hack_parser import format_line, split_C, get_command_type, A_is_symbol
import unittest

class TestParser(unittest.TestCase):
    def test_remove_comment(self):
        self.assertEqual(format_line("// test a basic comment is removed and would be ignored"), "")
        self.assertEqual(format_line("D=M+1  // test comments after some code is removed"), "D=M+1")
        self.assertEqual(format_line("     (LOOP)   // test loop label is left"), "(LOOP)")
        

    def test_remove_whitespace(self):
        self.assertEqual(format_line("D          =     M +         1"), "D=M+1")

    def test_ignore(self):
        lines = ["// This line is a comment", "", "         "]
        skipped = True
        for line in lines:
            line = format_line(line)
            if len(line) == 0: continue
            skipped = False
        self.assertEqual(skipped, True)

    def test_label_skip(self):
        lines = ["(LOOP)", "     (PAINT)", "(APPLE)    // This is a label   "]
        skipped = False
        for line in lines:
            line = format_line(line)
            if len(line) == 0: skipped = True
        self.assertEqual(skipped, False)

    def test_general(self):
        self.assertEqual(format_line("@1234"), "@1234")

    def test_get_command_type(self):
        self.assertEqual(get_command_type("@1234"), "A_COMMAND")
        self.assertEqual(get_command_type("D=M+1"), "C_COMMAND")
        self.assertEqual(get_command_type("0;JMP"), "C_COMMAND")
        self.assertEqual(get_command_type("(LOOP)"), "L_COMMAND")

    def test_split_C(self):
        self.assertEqual(split_C("a=b;c"), ("a", "b", "c"))
        self.assertEqual(split_C("a=b"), ("a", "b", "null"))
        self.assertEqual(split_C("b;c"), ("null", "b", "c"))

    def test_A_is_symbol(self):
        self.assertEqual(A_is_symbol("rabbit"), True)
        self.assertEqual(A_is_symbol("123"), False)
        self.assertEqual(A_is_symbol("2"), False)
        self.assertEqual(A_is_symbol("0"), False)

if __name__== '__main__':
    unittest.main()