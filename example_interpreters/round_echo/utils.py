import sys
import signal
import time

class TimeoutException(Exception):
    pass

def Timeout(signum, frame):
    Log("Read time-out")

    time.sleep(1)

    raise TimeoutException()

def Init():
    assert len(sys.argv) >= 3, "not enough arguments"

    n, i = int(sys.argv[1]), int(sys.argv[2])

    assert 0 <= i < n, "index not understood"

    return i, n, sys.argv[3:]

def Log(content):
    print(f"Log,{ content }")

def Send(i, content):
    print(f"Send,{ i },{ content }")

def Read(i, timeout = -1, handler = Timeout):
    if timeout > 0:
        signal.signal(signal.SIGALRM, handler) #link the SIGALRM signal to the handler
        signal.alarm(timeout) #create an alarm of timeout second

    resp = input(f"Req,{ i }\n")

    if timeout > 0:
        signal.alarm(0) #reinitiate the alarm

    return resp

def Reset(i):
    print(f"Reset,{ i }")

def Exit():
    print("Exit")
