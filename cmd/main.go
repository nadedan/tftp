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

	emptyData := bytes.NewReader([]byte{})
	err = tftp.Put("172.19.0.71", "empty", emptyData, tftp.WithBlocksize(6))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tftp.Put("172.19.0.71", "hello.txt", f, tftp.WithBlocksize(512), tftp.WithTimeoutSeconds(1))
	if err != nil {
		fmt.Println(err)
		return
	}
}
