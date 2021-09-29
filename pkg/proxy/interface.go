package proxy

type Proxy interface {
	Start(stopCh chan struct{}) error
}
