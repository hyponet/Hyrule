package transfer

type Transfer interface {
	SayHello(cup *PaperCup) (*PaperCup, error)
	Accept() (*PaperCup, error)
	Ping(cup *PaperCup) (*PaperCup, error)
}
