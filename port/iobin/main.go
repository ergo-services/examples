package main

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"ergo.services/ergo/lib"
)

func main() {
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if n == 0 {
					// closed connection
					return
				}

				fmt.Println("unable to read from stdin: ", err)
				return
			}
			if n == 0 {
				continue
			}
			fmt.Println("GOT:", buf[:n])
		}

	}()

	buf := make([]byte, 64)
	// header [...llll] payload with len PL...
	// where llll = PL+7
	for {
		s := lib.RandomString(rand.IntN(20))
		// fmt.Printf("random string: %s\n", s)
		l := 7 + len(s)
		binary.BigEndian.PutUint32(buf[3:7], uint32(l))
		copy(buf[7:], s)
		os.Stdout.Write(buf[:l])
		fmt.Fprintf(os.Stderr, "(iobin via stderr) sent %q\n", s)
		time.Sleep(time.Second)
	}
}
