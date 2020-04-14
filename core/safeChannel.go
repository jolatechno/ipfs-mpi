package core

import (
  "sync"
)

type SafeChannelBool struct {
	C chan bool
	Ended bool
	Mutex sync.Mutex
}

func NewChannelBool() *SafeChannelBool {
	return &SafeChannelBool{C: make(chan bool)}
}

func (sc *SafeChannelBool) SafeClose() {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
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

type SafeChannelString struct {
	C chan string
	Ended bool
	Mutex sync.Mutex
}

func NewChannelString() *SafeChannelString {
	return &SafeChannelString{C: make(chan string)}
}

func (sc *SafeChannelString) SafeClose() {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
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

type SafeChannelError struct {
	C chan error
	Ended bool
	Mutex sync.Mutex
}

func NewChannelError() *SafeChannelError {
	return &SafeChannelError{C: make(chan error)}
}

func (sc *SafeChannelError) SafeClose() {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
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

type SafeChannelInt struct {
	C chan int
	Ended bool
	Mutex sync.Mutex
}

func NewChannelInt() *SafeChannelInt {
	return &SafeChannelInt{C: make(chan int)}
}

func (sc *SafeChannelInt) SafeClose() {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
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

type SafeChannelMessage struct {
	C chan Message
	Ended bool
	Mutex sync.Mutex
}

func NewChannelMessage() *SafeChannelMessage {
	return &SafeChannelMessage{C: make(chan Message)}
}

func (sc *SafeChannelMessage) SafeClose() {
	sc.Mutex.Lock()
	defer sc.Mutex.Unlock()
	if !sc.Ended {
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
