package utils

import (
	"io"
)

type HalfCloser interface {
	io.ReadWriteCloser

	CloseWrite() error
	CloseRead() error
}

func CopyAndClose(dst, src HalfCloser) error {
	defer func() {
		_ = dst.CloseWrite()
		_ = src.CloseRead()
	}()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func CopyOnly(dst io.Writer, src io.Reader) error {
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
