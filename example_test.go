package gopc_test

import (
	"bytes"
	"log"

	"github.com/qmuntal/gopc"
)

func ExampleWriter() {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new OPC archive.
	w := gopc.NewWriter(buf)

	// Create a new OPC part.
	name, err := gopc.NormalizePartName("docs\\readme.txt")
	if err != nil {
		log.Fatal(err)
	}
	part, err := w.Create(name, "text/plain", gopc.CompressionNormal)
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the part.
	_, err = part.Write([]byte("This archive contains some text files."))
	if err != nil {
		log.Fatal(err)
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
}
