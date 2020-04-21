import utils
import time

if __name__ == "__main__":
    i, n, args = utils.Init()

    assert n == 2, "size not understood"

    if i == 0:
        if len(args) >= 1:
            msg = " ".join(args)
        else:
            msg = "sending"

        utils.Log(f"sending \"{ msg }\"")
        utils.Send(1, msg)
        resp = utils.Read(1, 5)
        utils.Log(f"responded \"{ resp }\"")

    else:
        msg = resp = utils.Read(0)

        time.sleep(2)

        utils.Log(f"responding \"{msg} {msg}\"")
        utils.Send(0, f"{msg} {msg}")
