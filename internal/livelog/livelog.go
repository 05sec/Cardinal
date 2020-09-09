package livelog

import (
	"context"
	"errors"
	"sync"
)

type Streamer struct {
	sync.Mutex

	streams map[string]*stream
}

var errStreamNotFound = errors.New("stream: not found")

// newStreamer returns a new in-memory log streamer.
func newStreamer() *Streamer {
	return &Streamer{
		streams: make(map[string]*stream),
	}
}

// Create adds a new log stream.
func (s *Streamer) Create(id string) error {
	s.Lock()
	s.streams[id] = newStream()
	s.Unlock()
	return nil
}

// Delete removes a log by id.
func (s *Streamer) Delete(id string) error {
	s.Lock()
	stream, ok := s.streams[id]
	if ok {
		delete(s.streams, id)
	}
	s.Unlock()
	if !ok {
		return errStreamNotFound
	}
	return stream.close()
}

// Write adds a new line into stream.
func (s *Streamer) Write(id string, line *Line) error {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()
	if !ok {
		return errStreamNotFound
	}
	return stream.write(line)
}

// Tail returns the end signal.
func (s *Streamer) Tail(ctx context.Context, id string) (<-chan *Line, <-chan error) {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()
	if !ok {
		return nil, nil
	}
	return stream.subscribe(ctx)
}

// Info returns the count of subscribers in each stream.
func (s *Streamer) Info() map[string]int {
	s.Lock()
	defer s.Unlock()
	info := map[string]int{}
	for id, stream := range s.streams {
		stream.Lock()
		info[id] = len(stream.sub)
		stream.Unlock()
	}
	return info
}
