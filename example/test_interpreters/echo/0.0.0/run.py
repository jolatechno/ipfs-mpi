import sys
import base64

if __name__ == "__main__":
    msg = sys.argv[1]
    splitted = msg.split(',')
    assert len(splitted) == 4, "message not well formatted"

    splitted[0], splitted[1], splitted[2] = str(base64.b64encode(splitted[3].encode("utf-8")), "utf8"), splitted[2], splitted[1]
    splitted[3] += str(base64.b64encode("echo".encode("utf-8")), "utf-8")
    print(",".join(splitted))
