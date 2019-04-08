package opc

import (
	"archive/zip"
	"io"
)

type zipFile struct {
	f *zip.File
}

func (zf *zipFile) Open() (io.ReadCloser, error) {
	return zf.f.Open()
}

func (zf *zipFile) Name() string {
	return zf.f.Name
}

func (zf *zipFile) Size() int {
	return int(zf.f.UncompressedSize64)
}

type zipArchive struct {
	r *zip.Reader
}

func newZipReader(r io.ReaderAt, size int64) (*zipArchive, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	return &zipArchive{zr}, nil
}

func (z *zipArchive) Files() []archiveFile {
	files := z.r.File
	ret := make([]archiveFile, len(files))
	for i := 0; i < len(files); i++ {
		ret[i] = &zipFile{files[i]}
	}
	return ret
}

func (z *zipArchive) RegisterDecompressor(method uint16, dcomp func(r io.Reader) io.ReadCloser) {
	z.r.RegisterDecompressor(method, dcomp)
}
