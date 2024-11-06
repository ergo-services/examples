package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	var i int

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
		}

	}()

	for {
		i++
		fmt.Fprintf(os.Stdout, "(iotxt via stdout) example TXT-%03d message\n", i)
		fmt.Fprintf(os.Stderr, "(iotxt via stderr) example ERR-%03d message\n", i)
		time.Sleep(time.Second)
	}
}
