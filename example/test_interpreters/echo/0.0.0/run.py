import sys
import binascii
import codecs

if __name__ == "__main__":
    msg = sys.argv[1]
    splitted = msg.split(',')
    assert len(splitted) == 4, "message not well formatted"

    splitted[0], splitted[1], splitted[2] = codecs.decode(splitted[3], "hex").decode("utf8"), splitted[2], splitted[1]
    splitted[3] += binascii.hexlify(b"echo").decode("utf-8")
    print(",".join(splitted))
