package transfer

type HttpTransfer struct {
}

var _ Transfer = &HttpTransfer{}

func (h HttpTransfer) SayHello(cup *PaperCup) (*PaperCup, error) {
	panic("implement me")
}

func (h HttpTransfer) Ping(cup *PaperCup) (*PaperCup, error) {
	panic("implement me")
}

func (h HttpTransfer) Accept() (*PaperCup, error) {
	panic("implement me")
}
