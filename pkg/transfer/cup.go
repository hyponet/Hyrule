package transfer

import (
	"github.com/Coderhypo/Hyrule/utils"
)

const (
	PacketHello     = "hello"
	PacketHeartBeat = "ping"
	PacketConnected = "connected"
	PacketError     = "error"
)

type PaperCup struct {
	Type    string
	Addr    string
	Message string
	Data    utils.HalfCloser
}
