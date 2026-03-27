package channels

import (
	"context"
	"log"
	"sync"
	"time"
)

type OutboundHandler func(msg *OutboundMessage)

type MessageBus struct {
	inboundQueue     chan *InboundMessage
	outboundMu       sync.RWMutex
	outboundHandlers []OutboundHandler
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inboundQueue: make(chan *InboundMessage, 100),
	}
}

func (b *MessageBus) PublishInbound(msg *InboundMessage) {
	msg.CreatedAt = time.Now()
	b.inboundQueue <- msg
	log.Printf("[Bus] inbound published: channel=%s, chat_id=%s, type=%s", msg.ChannelName, msg.ChatID, msg.MsgType)
}

func (b *MessageBus) GetInbound(ctx context.Context) (*InboundMessage, error) {
	select {
	case msg := <-b.inboundQueue:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *MessageBus) SubscribeOutbound(handler OutboundHandler) {
	b.outboundMu.Lock()
	defer b.outboundMu.Unlock()
	b.outboundHandlers = append(b.outboundHandlers, handler)
}

func (b *MessageBus) UnsubscribeOutbound(handler OutboundHandler) {
	b.outboundMu.Lock()
	defer b.outboundMu.Unlock()
	for i, h := range b.outboundHandlers {
		if &h == &handler {
			b.outboundHandlers = append(b.outboundHandlers[:i], b.outboundHandlers[i+1:]...)
			break
		}
	}
}

func (b *MessageBus) PublishOutbound(msg *OutboundMessage) {
	msg.CreatedAt = time.Now()
	b.outboundMu.RLock()
	defer b.outboundMu.RUnlock()
	log.Printf("[Bus] outbound dispatching: channel=%s, chat_id=%s, handlers=%d", msg.ChannelName, msg.ChatID, len(b.outboundHandlers))
	for _, handler := range b.outboundHandlers {
		go handler(msg)
	}
}
