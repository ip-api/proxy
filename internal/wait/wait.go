package wait

type WaitForChannels map[chan struct{}]struct{}

func New() WaitForChannels {
	return make(map[chan struct{}]struct{})
}

func (w WaitForChannels) Add(b chan struct{}) {
	w[b] = struct{}{}
}

func (w WaitForChannels) Wait() {
	for c := range w {
		<-c
	}
}
