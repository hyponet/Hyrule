package ganon

import (
	"context"
	"github.com/Coderhypo/Hyrule/pkg/transfer"
	"github.com/Coderhypo/Hyrule/utils"
	"go.uber.org/zap"
	"net"
)

type LocalCastle struct {
	transfer.Transfer
	logger *zap.SugaredLogger
}

func (l *LocalCastle) Start(stopCh chan struct{}) error {
	l.logger.Info("server start")

	for {
		select {
		case <-stopCh:
			return nil
		default:
			l.logger.Debug("wait new cup")
		}

		cup, err := l.Accept()
		if err != nil {
			l.logger.Errorw("accept got cup failed", "err", err.Error())
			continue
		}
		go l.handler(context.TODO(), cup)
	}
}

func (l *LocalCastle) handler(ctx context.Context, cup *transfer.PaperCup) {
	l.logger.Debugw("handle new connect", "addr", cup.Addr)
	conn, err := net.Dial("tcp", cup.Addr)
	if err != nil {
		l.logger.Errorw("connect remote server failed", "addr", cup.Addr, "err", err.Error())
		return
	}

	connHalfCloser, ok := conn.(utils.HalfCloser)
	if !ok {
		_ = cup.Data.Close()
		_ = conn.Close()
		l.logger.Errorw("not tcp conn")
		return
	}
	go func() {
		if err := utils.CopyAndClose(cup.Data, connHalfCloser); err != nil {
			l.logger.Errorw("copy conn data to cup failed", "err", err.Error())
		}
		l.logger.Debug("copy conn data to cup finish")
	}()
	go func() {
		if err := utils.CopyAndClose(connHalfCloser, cup.Data); err != nil {
			l.logger.Errorw("copy cup data to conn failed", "err", err.Error())
		}
		l.logger.Debug("copy cup data to conn finish")
	}()
	l.logger.Debugw("handler finish")
}

func NewLocalCastle(dep Dep) Castle {
	return &LocalCastle{
		Transfer: dep.Transfer,
		logger:   utils.NewLogger("localCastle"),
	}
}
