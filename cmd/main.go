package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nadedan/tftp"
)

func main() {
	f, err := os.Open("../testdata/some.txt")
	if err != nil {
		panic(err)
	}

	tmp := ""
	fmt.Printf("Press Enter when a TFTP server is running on localhost\n and does not have files named 'empty' or 'hello.txt'")
	fmt.Scanln(&tmp)
	fmt.Println()

	fmt.Printf("PUTting a 0-byte file to 'empty'\n")
	emptyData := bytes.NewReader([]byte{})
	err = tftp.Put("127.0.0.1", "empty", emptyData, tftp.WithBlocksize(6))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("PUTting a 1000-line 'Hello, World!' file to 'hello.txt'\n")
	err = tftp.Put("127.0.0.1", "hello.txt", f, tftp.WithBlocksize(512), tftp.WithTimeoutSeconds(1))
	if err != nil {
		fmt.Println(err)
		return
	}
}
