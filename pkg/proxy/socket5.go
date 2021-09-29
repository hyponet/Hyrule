package proxy

const (
	TypeSocket5 = "socket5"
)

type Socket5 struct{}

var _ Proxy = &Socket5{}

func (s Socket5) Start(stopCh chan struct{}) error {
	panic("implement me")
}
