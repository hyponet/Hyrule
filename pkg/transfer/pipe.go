package transfer

import (
	"bytes"
	"errors"
	"github.com/Coderhypo/Hyrule/utils"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

const (
	pipeDefaultCap    = 2048
	pipeConnectTimout = time.Second * 10
)

var pipeQ chan *PaperCup

type PipeTransfer struct {
	logger *zap.SugaredLogger
}

var _ Transfer = &PipeTransfer{}

func (p *PipeTransfer) SayHello(cup *PaperCup) (*PaperCup, error) {
	peer, local := newPipe()
	cup.Data = peer
	timout := time.NewTimer(pipeConnectTimout)
	defer timout.Stop()
	select {
	case pipeQ <- cup:
		p.logger.Debug("peer connected")
	case <-timout.C:
		return nil, errors.New("timeout")
	}
	return &PaperCup{
		Type:    PacketConnected,
		Addr:    cup.Addr,
		Message: "",
		Data:    local,
	}, nil
}

func (p *PipeTransfer) Ping(cup *PaperCup) (*PaperCup, error) {
	panic("implement me")
}

func (p *PipeTransfer) Accept() (*PaperCup, error) {
	newCup := <-pipeQ
	return newCup, nil
}

func NewPipeTransfer() Transfer {
	pipeQ = make(chan *PaperCup)
	return &PipeTransfer{
		logger: utils.NewLogger("pipeTransfer"),
	}
}

type pipe struct {
	readBuf       *halfCloseBuffer
	readCount     int64
	writeBuf      *halfCloseBuffer
	writeCount    int64
	rCanReadCond  *sync.Cond
	rCanWriteCond *sync.Cond
	wCanReadCond  *sync.Cond
	wCanWriteCond *sync.Cond
}

func (q *pipe) Read(p []byte) (n int, err error) {
	q.rCanReadCond.L.Lock()
	defer q.rCanReadCond.L.Unlock()
	for q.readBuf.Len() == 0 {
		q.rCanReadCond.Wait()
	}
	if q.readBuf.Len() == 0 {
		return 0, io.EOF
	}
	n, err = q.readBuf.Read(p)
	q.rCanWriteCond.Broadcast()
	q.readCount += int64(n)
	return
}

func (q *pipe) Write(p []byte) (n int, err error) {
	q.wCanWriteCond.L.Lock()
	defer q.wCanWriteCond.L.Unlock()
	for len(p) > 0 {
		onceWrite := 0
		canWrite := pipeDefaultCap - q.writeBuf.Len()
		for canWrite == 0 {
			q.wCanWriteCond.Wait()
			canWrite = pipeDefaultCap - q.writeBuf.Len()
		}
		if canWrite > len(p) {
			canWrite = len(p)
		}
		onceWrite, err = q.writeBuf.Write(p[:canWrite])
		if err != nil {
			return
		}
		q.wCanReadCond.Broadcast()
		p = p[onceWrite:]
		n += onceWrite
	}
	q.writeCount += int64(n)
	return
}

func (q *pipe) CloseWrite() error {
	return q.writeBuf.CloseWrite()
}

func (q *pipe) CloseRead() error {
	return q.readBuf.CloseRead()
}

func (q *pipe) Close() error {
	_ = q.readBuf.Close()
	_ = q.writeBuf.Close()
	return nil
}

type halfCloseBuffer struct {
	buf         *bytes.Buffer
	readClosed  bool
	writeClosed bool
}

func (h *halfCloseBuffer) Read(p []byte) (n int, err error) {
	if h.buf.Len() == 0 && h.readClosed {
		return 0, io.ErrClosedPipe
	}
	return h.buf.Read(p)
}

func (h *halfCloseBuffer) Write(p []byte) (n int, err error) {
	if h.writeClosed {
		return 0, io.ErrClosedPipe
	}
	return h.buf.Write(p)
}
func (h *halfCloseBuffer) Len() int {
	return h.buf.Len()
}

func (h *halfCloseBuffer) Close() error {
	h.readClosed = true
	h.writeClosed = true
	return nil
}

func (h *halfCloseBuffer) CloseWrite() error {
	h.writeClosed = true
	return nil
}

func (h *halfCloseBuffer) CloseRead() error {
	h.readClosed = true
	return nil
}

func newPipe() (*pipe, *pipe) {
	pipeBuf1 := &halfCloseBuffer{buf: bytes.NewBuffer([]byte{})}
	pipeBuf2 := &halfCloseBuffer{buf: bytes.NewBuffer([]byte{})}

	buf1Lock := &sync.Mutex{}
	buf1CanRead := sync.NewCond(buf1Lock)
	buf1CanWrite := sync.NewCond(buf1Lock)

	buf2Lock := &sync.Mutex{}
	buf2CanRead := sync.NewCond(buf2Lock)
	buf2CanWrite := sync.NewCond(buf2Lock)

	peer := &pipe{
		readBuf:       pipeBuf1,
		writeBuf:      pipeBuf2,
		rCanReadCond:  buf1CanRead,
		rCanWriteCond: buf1CanWrite,
		wCanReadCond:  buf2CanRead,
		wCanWriteCond: buf2CanWrite,
	}
	local := &pipe{
		readBuf:       pipeBuf2,
		writeBuf:      pipeBuf1,
		rCanReadCond:  buf2CanRead,
		rCanWriteCond: buf2CanWrite,
		wCanReadCond:  buf1CanRead,
		wCanWriteCond: buf1CanWrite,
	}
	return peer, local
}
