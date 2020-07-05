package wait

type BlockAndError struct {
	C   chan struct{}
	Err error
}

type WaitForBlockAndError map[*BlockAndError]struct{}

func New() WaitForBlockAndError {
	return make(map[*BlockAndError]struct{})
}

func (w WaitForBlockAndError) Add(b *BlockAndError) {
	w[b] = struct{}{}
}

func (w WaitForBlockAndError) Wait() error {
	for b := range w {
		if b.Err != nil {
			return b.Err
		}

		if b.C != nil {
			<-b.C
		}
	}

	return nil
}
