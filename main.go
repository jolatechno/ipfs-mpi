package main

import (
  "fmt"
  "context"
  "bufio"
  "os"
  "time"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/jolatechno/ipfs-mpi/core"
)

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func main(){
  ctx := context.Background()

  h, err := core.NewHost(ctx)
  if err != nil {
    panic(err)
  }

  h.SetStreamHandler(protocol.ID("protocol/1.0.0"), handleStream)

  core.StartDiscovery(ctx, h, "rendezvous")

  fmt.Println("Our adress is: ", h.ID())
  for _, addr := range h.Addrs() {
    fmt.Println("Swarm listenning on: ", addr)
  }

  l := 0
  for {
    time.Sleep(time.Second)
    if len(h.Peerstore().Peers()) > l {
      fmt.Println("New peer ", h.Peerstore().Peers()[0])
      l++
    }
  }

  select {}
}
