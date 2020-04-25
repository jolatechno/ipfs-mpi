import utils

if __name__ == "__main__":
    i, n, args = utils.Init()

    assert n >= 2, "size not understood"


    if i == 0:
        if len(args) >= 1:
            msg = " ".join(args)
        else:
            msg = "sending"

        utils.Log(f"sending \"{ msg }\"")
        utils.Send(1, msg)
        resp = utils.Read(1, 20)
        utils.Log(f"responded \"{ resp }\"")

    elif i == n - 1:
        msg = utils.Read(i - 1)
        utils.Log(f"echo \"{ msg }\"")
        utils.Log(f"sending \"echo { msg }\" to { i - 1 }")
        utils.Send(i - 1, f"echo { msg }")

    else:
        msg = utils.Read(i - 1)
        utils.Log(f"sending \"{ msg }\" to { i + 1 }")
        utils.Send((i + 1), msg)

        msg = utils.Read(i + 1)
        utils.Log(f"sending \"{ msg }\" to { i - 1 }")
        utils.Send(i - 1, msg)
