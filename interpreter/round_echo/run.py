import sys
import base64

if __name__ == "__main__":
    n, i = int(sys.argv[1]), int(sys.argv[2])

    assert n >= 2, "size not understood"
    assert 0 <= i < n, "index not understood"

    if i == 0:
        if len(sys.argv) == 4:
            msg = sys.argv[3]
        else:
            msg = "sending"

        print("Log," + msg)
        print(f"Send,1,{msg}")

        resp = input(f"Req,{n - 1}\n")
        print(f"Log,{n - 1} responded: {resp}")
    else:
        msg = input(f"Req,{i - 1}\n")
        print(f"Send,{(i + 1)%n},{msg}")
