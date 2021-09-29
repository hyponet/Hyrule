package proxy

import "github.com/Coderhypo/Hyrule/pkg/transfer"

type Config struct {
	HttpAddr string `json:"http_addr"`
	HttpPort int32  `json:"http_port"`
}

type Dep struct {
	Config   Config
	Transfer transfer.Transfer
}
