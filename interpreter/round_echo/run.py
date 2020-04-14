import sys
import signal

class TimeoutException(Exception):
    pass

def Timeout(signum, frame):
    print("Log,Timed-out")
    raise TimeoutException()

if __name__ == "__main__":
    n, i = int(sys.argv[1]), int(sys.argv[2])

    assert n >= 2, "size not understood"
    assert 0 <= i < n, "index not understood"

    timeout = min(60, 2*n) # timeout duration

    if i == 0:
        if len(sys.argv) == 4:
            msg = sys.argv[3]
        else:
            msg = "sending"

        print("Log," + msg)
        print(f"Send,1,{msg}")

        signal.signal(signal.SIGALRM, Timeout) #link the SIGALRM signal to the handler
        signal.alarm(timeout) #create an alarm of timeout second

        resp = input(f"Req,{n - 1}\n")
        print(f"Log,{n - 1} responded: {resp}")

        signal.alarm(0) #reinitiate the alarm
    else:
        msg = input(f"Req,{i - 1}\n")
        if i == n - 1:
            print(f"Send,0,{msg}")
        else:
            print(f"Send,{i + 1},{msg}")
