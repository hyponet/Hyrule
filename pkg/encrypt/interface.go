package encrypt

import "io"

type Interface interface {
	Encode(data io.ReadCloser) (io.ReadCloser, error)
	Decode(data io.ReadCloser) (io.ReadCloser, error)
}
