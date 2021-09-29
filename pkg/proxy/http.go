package proxy

import (
	"context"
	"fmt"
	"github.com/Coderhypo/Hyrule/pkg/transfer"
	"github.com/Coderhypo/Hyrule/utils"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type HttpProxy struct {
	transfer.Transfer

	HttpAddr string
	HttpPort int32
	logger   *zap.SugaredLogger
}

var _ Proxy = &HttpProxy{}

func (h *HttpProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	peer, err := h.Transfer.SayHello(&transfer.PaperCup{Type: transfer.PacketHello, Addr: getRemoteAddr(request.URL)})
	if err != nil {
		h.logger.Errorw("say hello failed", "err", err.Error())
		h.handleErr(writer, err.Error())
		return
	}

	if peer.Type != transfer.PacketConnected {
		h.logger.Errorw("not got connect", "message", peer.Message)
		h.handleErr(writer, "got peer conn failed")
		return
	}

	if request.Method == http.MethodConnect {
		h.logger.Debugw("handle connect method", "host", request.Host)
		h.handleHttps(writer, request, peer)
		return
	}
	h.handleHttp(writer, request, peer)
}

func (h *HttpProxy) handleHttps(writer http.ResponseWriter, request *http.Request, cup *transfer.PaperCup) {
	hij, ok := writer.(http.Hijacker)
	if !ok {
		h.logger.Error("do not support hijacking")
		h.handleErr(writer, "do not support hijacking")
		return
	}

	conn, _, err := hij.Hijack()
	if err != nil {
		h.logger.Errorw("get hijack conn failed", "err", err.Error())
		h.handleErr(writer, "got conn failed")
		return
	}
	if _, err = fmt.Fprint(conn, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		h.logger.Errorw("send conn established failed", "err", err.Error())
	}
	h.logger.Debug("send conn established")
	go h.transfer(conn, cup)

}
func (h *HttpProxy) handleHttp(writer http.ResponseWriter, request *http.Request, cup *transfer.PaperCup) {
	hij, ok := writer.(http.Hijacker)
	if !ok {
		h.logger.Error("do not support hijacking")
		h.handleErr(writer, "do not support hijacking")
		return
	}

	conn, _, err := hij.Hijack()
	if err != nil {
		h.logger.Errorw("get hijack conn failed", "err", err.Error())
		h.handleErr(writer, "got conn failed")
		return
	}
	reqRaw, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		h.handleErr(writer, "send request failed")
	}
	if _, err = cup.Data.Write(reqRaw); err != nil {
		h.logger.Errorw("send http request failed", "err", err.Error())
	}
	go h.transfer(conn, cup)
}

func (h *HttpProxy) transfer(conn net.Conn, cup *transfer.PaperCup) {
	logCopyErr := func(err error) {
		if err != nil {
			h.logger.Errorw("copy data error", "err", err.Error())
		}
	}

	connHalfCloser, ok := conn.(utils.HalfCloser)
	if ok {
		h.logger.Debug("conn is half closer")
		go func() {
			logCopyErr(utils.CopyAndClose(cup.Data, connHalfCloser))
		}()
		go func() {
			logCopyErr(utils.CopyAndClose(connHalfCloser, cup.Data))
		}()
		return
	}

	go func() {
		h.logger.Debug("conn not half closer")
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			logCopyErr(utils.CopyOnly(conn, cup.Data))
			logCopyErr(utils.CopyOnly(cup.Data, conn))
		}()
	}()
}

func (h *HttpProxy) handleErr(writer http.ResponseWriter, errMsg string) {
	_, _ = writer.Write([]byte(errMsg))
}

func (h *HttpProxy) Start(stopCh chan struct{}) error {
	lisAddr := fmt.Sprintf("%s:%d", h.HttpAddr, h.HttpPort)
	server := http.Server{Addr: lisAddr, Handler: h}

	go func() {
		h.logger.Infof("start http proxy on %s", lisAddr)
		if err := server.ListenAndServe(); err != nil {
			h.logger.Errorw("start http proxy failed", "err", err.Error())
			panic(err)
		}
	}()

	<-stopCh
	return server.Shutdown(context.TODO())
}

func NewHttpProxy(dep Dep) Proxy {
	cfg := dep.Config
	return &HttpProxy{
		Transfer: dep.Transfer,

		HttpAddr: cfg.HttpAddr,
		HttpPort: cfg.HttpPort,
		logger:   utils.NewLogger("httpProxy"),
	}
}

func getRemoteAddr(u *url.URL) string {
	result := u.Host
	if u.Port() != "" {
		return result
	}
	switch u.Scheme {
	case "http":
		result += ":80"
	case "https":
		result += ":443"
	}
	return result
}
