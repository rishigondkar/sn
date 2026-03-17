package consumer

import (
	"context"
	"sync"
)

// ChannelSource is a MessageSource that reads messages from a channel (POC/testing).
// For production, replace with an SQS/Kafka implementation.
type ChannelSource struct {
	ch   chan *IncomingMessage
	once sync.Once
}

// NewChannelSource creates a source that reads from the given channel.
// The channel is closed when the source is closed.
func NewChannelSource(buf int) *ChannelSource {
	return &ChannelSource{ch: make(chan *IncomingMessage, buf)}
}

// Receive blocks until a message is available or ctx is cancelled.
func (s *ChannelSource) Receive(ctx context.Context) (*IncomingMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-s.ch:
		if !ok {
			return nil, ctx.Err()
		}
		return msg, nil
	}
}

// Send enqueues a message (for tests or POC drivers). Body is the raw JSON bytes.
func (s *ChannelSource) Send(body []byte) {
	ack := func() {}
	nack := func() {}
	s.ch <- &IncomingMessage{Body: body, Ack: ack, Nack: nack}
}

// Close closes the channel; Receive will return after drain.
func (s *ChannelSource) Close() {
	s.once.Do(func() { close(s.ch) })
}
