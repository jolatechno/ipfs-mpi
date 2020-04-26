import utils
import time
import random

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

        time.sleep(2)

        utils.Log(f"sending \"echo { resp }\"")
        utils.Send(1, f"echo { resp }")
        secondresp = utils.Read(1, 5)
        utils.Log(f"responded \"{ secondresp }\"")

    else:
        msg = resp = utils.Read(0)
        utils.Log(f"responding \"echo {msg}\"")
        utils.Send(0, f"echo {msg}")

        assert random.random() < 0.5, "failed on purpose"

        msg = resp = utils.Read(0)
        utils.Log(f"responding \"echo {msg}\" for the second time")
        utils.Send(0, f"echo {msg}")
