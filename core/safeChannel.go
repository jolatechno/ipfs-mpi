package core

import (
  "sync"
)

func NewChannelBool() *SafeChannelBool {
	return &SafeChannelBool{C: make(chan bool)}
}

type SafeChannelBool struct {
	C chan bool
	Ended bool
	Mutex sync.Mutex
}

func (sc *SafeChannelBool)SafeClose(clear bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    if clear {
      for len(sc.C) > 0 {
        <- sc.C
      }
    }
		close(sc.C)
		sc.Ended = true
	}
}

func (sc *SafeChannelBool) Send(t bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    sc.C <- t
  }
}

//-------

func NewChannelString() *SafeChannelString {
	return &SafeChannelString{C: make(chan string)}
}

type SafeChannelString struct {
	C chan string
	Ended bool
	Mutex sync.Mutex
}

func (sc *SafeChannelString)SafeClose(clear bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    if clear {
      for len(sc.C) > 0 {
        <- sc.C
      }
    }
		close(sc.C)
		sc.Ended = true
	}
}

func (sc *SafeChannelString) Send(str string) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    sc.C <- str
  }
}

//-------

func NewChannelError() *SafeChannelError {
	return &SafeChannelError{C: make(chan error)}
}

type SafeChannelError struct {
	C chan error
	Ended bool
	Mutex sync.Mutex
}

func (sc *SafeChannelError)SafeClose(clear bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    if clear {
      for len(sc.C) > 0 {
        <- sc.C
      }
    }
		close(sc.C)
		sc.Ended = true
	}
}

func (sc *SafeChannelError) Send(err error) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    sc.C <- err
  }
}

//-------

func NewChannelInt() *SafeChannelInt {
	return &SafeChannelInt{C: make(chan int)}
}

type SafeChannelInt struct {
	C chan int
	Ended bool
	Mutex sync.Mutex
}

func (sc *SafeChannelInt)SafeClose(clear bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    if clear {
      for len(sc.C) > 0 {
        <- sc.C
      }
    }
		close(sc.C)
		sc.Ended = true
	}
}

func (sc *SafeChannelInt) Send(i int) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    sc.C <- i
  }
}

//-------

func NewChannelMessage() *SafeChannelMessage {
	return &SafeChannelMessage{C: make(chan Message)}
}

type SafeChannelMessage struct {
	C chan Message
	Ended bool
	Mutex sync.Mutex
}

func (sc *SafeChannelMessage)SafeClose(clear bool) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    if clear {
      for len(sc.C) > 0 {
        <- sc.C
      }
    }
		close(sc.C)
		sc.Ended = true
	}
}

func (sc *SafeChannelMessage) Send(msg Message) {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
    sc.C <- msg
  }
}
