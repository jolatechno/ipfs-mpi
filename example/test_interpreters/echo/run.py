import sys
import base64

if __name__ == "__main__":
    n, i = int(sys.argv[1]), int(sys.argv[2])

    assert n == 2, "size not understood"
    assert 0 <= i <= 2, "index not understood"

    if i == 0:
        if len(sys.argv) == 4:
            msg = sys.argv[3]
        else:
            msg = "sending"
        print("Log," + msg)
        print("1," + msg)

        resp = input("Req,1\n")
        print("Log,responded: " + resp)
    else:
        msg = input("Req,0\n")
        print("0," + msg + " " + msg)
