package ganon

type Castle interface {
	Start(stopCh chan struct{}) error
}
