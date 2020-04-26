import utils
import time

if __name__ == "__main__":
    i, n, args = utils.Init()

    assert n >= 2, "size not understood"


    if i == 0:
        if len(args) >= 1:
            msg = " ".join(args)
        else:
            msg = "sending"

        utils.Log(f"sending \"{ msg }\"")

        for j in range(1, n):
            utils.Send(j, msg)

        for j in range(1, n):
            resp = utils.Read(j, 5)
            utils.Log(f"{ j } responded \"{ resp }\"")

    else:
        msg = utils.Read(0)

        time.sleep(2)

        utils.Log(f"responding \"echo { msg }\"")
        utils.Send(0, f"echo {msg}")
