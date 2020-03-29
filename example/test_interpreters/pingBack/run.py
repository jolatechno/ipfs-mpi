import sys
import base64

if __name__ == "__main__":
    Origin = sys.argv[1]
    Pid = int(sys.argv[2])

    message = input("Req;" + Origin + '\n')

    splitted = message.split(';')

    assert len(splitted) == 2, "message not well formatted"
    assert splitted[0] == "Msg", "header not understood"

    splitted = splitted[1].split(',')
    assert len(splitted) == 6, "message not well formatted"

    splitted[3], splitted[4] = splitted[4], splitted[3]
    print("Msg;" + ",".join(splitted) + "\n")
