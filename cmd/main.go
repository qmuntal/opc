package main

import (
	"log"
	"os"

	"github.com/qmuntal/gopc"
)

func main() {
	// Create a buffer to write our archive to.
	fl, err := os.Create("./carlota.upc")
	if err != nil {
		log.Fatal(err)
	}
	defer fl.Close()

	// Create a new OPC archive.
	w := gopc.NewWriter(fl)

	// Create a new OPC part.
	p := &gopc.Part{Name: gopc.NormalizePartName("docs\\readme.txt"), ContentType: "text/plain"}
	part1, err := w.CreatePart(p, gopc.CompressionNormal)
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the part.
	_, err = part1.Write([]byte("This archive contains some text files."))
	if err != nil {
		log.Fatal(err)
	}

	p.Relationships = append(p.Relationships, &gopc.Relationship{})

	// Create a new OPC part.
	_, err = w.Create(gopc.NormalizePartName("hello.txt"), "text/plain")
	if err != nil {
		log.Fatal(err)
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
}
