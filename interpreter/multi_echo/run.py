import sys
import signal

timeout = 5 # timeout duration

class TimeoutException(Exception):
    pass

def Timeout(signum, frame):
    print("Log,Timed-out")
    raise TimeoutException()

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
        for j in range(1, n):
            print(f"Send,{j},{msg}")

        for j in range(1, n):
            signal.signal(signal.SIGALRM, Timeout) #link the SIGALRM signal to the handler
            signal.alarm(timeout) #create an alarm of timeout second

            resp = input(f"Req,{j}\n")
            print(f"Log,{j} responded: {resp}")

            signal.alarm(0) #reinitiate the alarm
            
    else:
        msg = input("Req,0\n")
        print(f"Send,0,{msg} {msg}")
