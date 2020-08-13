package livelog

import "sync"

type subscriber struct {
	sync.Mutex

	handler      chan *Line
	closeChannel chan struct{}
	closed       bool
}

func (s *subscriber) send(line *Line) {
	select {
	case <-s.closeChannel:
	case s.handler <- line:
	default:

	}
}

func (s *subscriber) close() {
	s.Lock()
	if !s.closed {
		close(s.closeChannel)
		s.closed = true
	}
	s.Unlock()
}
