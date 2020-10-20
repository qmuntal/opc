package opc_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/qmuntal/opc"
)

func ExampleWriter() {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new OPC archive.
	w := opc.NewWriter(buf)

	// Create a new OPC part.
	name := opc.NormalizePartName("docs\\readme.txt")
	part, err := w.Create(name, "text/plain")
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the part.
	_, err = part.Write([]byte("This archive contains some text files."))
	if err != nil {
		log.Fatal(err)
	}

	// Make sure to check the error on Close.
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleReader() {
	r, err := opc.OpenReader("testdata/component.3mf")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.Files {
		fmt.Printf("Contents of %s with type %s :\n", f.Name, f.ContentType)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.CopyN(os.Stdout, rc, 68)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		fmt.Println()
	}
}

func ExampleNewWriterFromReader() {
	r, err := opc.OpenReader("testdata/component.3mf")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	w, err := opc.NewWriterFromReader(buf, r.Reader)
	if err != nil {
		log.Fatal(err)
	}
	name := opc.NormalizePartName("docs\\readme.txt")
	part, err := w.Create(name, "text/plain")
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the part.
	_, err = part.Write([]byte("This archive contains some text files."))
	if err != nil {
		log.Fatal(err)
	}

	// Make sure to check the error on Close.
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
}

