package util

import (
	"bytes"
	"compress/gzip"
	"io"

	"golang.org/x/text/transform"
)

type Zipper struct {
	r io.Reader
	c io.Closer
	w *gzip.Writer
	b bytes.Buffer
}

func (z *Zipper) Read(p []byte) (n int, err error) {
	return z.r.Read(p)
}
func (z *Zipper) Close() error {
	return z.c.Close()
}
func (z *Zipper) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if z.w == nil {
		z.w = gzip.NewWriter(&z.b)
	}
	nSrc, err = z.w.Write(src)
	if err != nil {
		return
	}
	if atEOF {
		_ = z.w.Close()
		nDst, err = z.b.Read(dst)
	} else {
		nDst, _ = z.b.Read(dst)
	}
	if z.b.Len() > 0 {
		err = transform.ErrShortDst
		return
	}
	return
}
func (z *Zipper) Reset() {
	z.w = nil
	z.b.Reset()
}

func NewZipper(r io.ReadCloser) *Zipper {
	z := new(Zipper)
	z.r = transform.NewReader(r, z)
	z.c = r
	return z
}
