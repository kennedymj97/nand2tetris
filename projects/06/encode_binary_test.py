import unittest

from encode_binary import encode_A, encode_C

class TestEncodeBinary(unittest.TestCase):
    def test_encode_A(self):
        self.assertEqual(encode_A('@7'), '0000000000000111')
        self.assertEqual(encode_A('@64'), '0000000001000000')

    def test_encode_C(self):
        self.assertEqual(encode_C("MD", "A-1", "null"), "1110110010011000")
        self.assertEqual(encode_C("null", "D", "JGE"), "1110001100000011")
        self.assertEqual(encode_C("AMD", "D&M", "null"), "1111000000111000")

if __name__ == "__main__":
    unittest.main()    