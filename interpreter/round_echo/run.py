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
        resp = utils.Read(n - 1, 10)
        utils.Log(f"{ n - 1 } responded: \"{ resp }\"")

    else:
        msg = resp = utils.Read(i - 1)
        utils.Send((i + 1)%n, msg)
