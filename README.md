# opc

[![PkgGoDev](https://pkg.go.dev/badge/github.com/qmuntal/opc)](https://pkg.go.dev/github.com/qmuntal/opc)
[![Build Status](https://travis-ci.com/qmuntal/opc.svg?branch=master)](https://travis-ci.com/qmuntal/opc)
[![Go Report Card](https://goreportcard.com/badge/github.com/qmuntal/opc)](https://goreportcard.com/report/github.com/qmuntal/opc)
[![codecov](https://coveralls.io/repos/github/qmuntal/opc/badge.svg)](https://coveralls.io/github/qmuntal/opc?branch=master)
[![codeclimate](https://codeclimate.com/github/qmuntal/opc/badges/gpa.svg)](https://codeclimate.com/github/qmuntal/opc)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

Package opc implements the ISO/IEC 29500-2, also known as the [Open Packaging Convention](https://en.wikipedia.org/wiki/Open_Packaging_Conventions).

The Open Packaging specification describes an abstract model and physical format conventions for the use of XML, Unicode, ZIP, and other openly available technologies and specifications to organize the content and resources of a document within a package.

The OPC is the foundation technology for many new file formats: .docx, .pptx, .xlsx, .3mf, .dwfx, ...

## Features
- [x] Package reader and writer
- [x] Package core properties and relationships
- [x] Part relationships
- [x] ZIP mapping
- [x] Package, relationships and parts validation against specs
- [ ] Part interleaved pieces
- [ ] Digital signatures

## Examples
### Write
```go
// Create a file to write our archive to.
f, _ := os.Create("example.xlsx")

// Create a new OPC archive.
w := opc.NewWriter(f)

// Create a new OPC part.
name := opc.NormalizePartName("docs\\readme.txt")
part, _ := w.Create(name, "text/plain")

// Write content to the part.
part.Write([]byte("This archive contains some text files."))

// Make sure to check the error on Close.
w.Close()
```

### Read
```go
r, _ := opc.OpenReader("testdata/test.xlsx")
defer r.Close()

// Iterate through the files in the archive,
// printing some of their contents.
for _, f := range r.Files {
  fmt.Printf("Contents of %s with type %s :\n", f.Name, f.ContentType)
  rc, _ := f.Open()
  io.CopyN(os.Stdout, rc, 68)
  rc.Close()
  fmt.Println()
}
```
