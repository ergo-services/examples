package main

import (
	"fmt"
	"os"
	"time"
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
			// fmt.Println("GOT:", string(buf[:n]))
		}

	}()

	i := 0
	for {
		i++
		fmt.Fprintf(os.Stdout, "(iotxt) example TXT-%03d message\n", i)
		fmt.Fprintf(os.Stderr, "(iotxt) example ERR-%03d message\n", i)
		time.Sleep(time.Second)
	}
}
