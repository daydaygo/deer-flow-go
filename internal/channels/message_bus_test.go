package channels

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMessageBusPublishInbound(t *testing.T) {
	bus := NewMessageBus()

	msg := &InboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		UserID:      "U123",
		Text:        "hello",
		MsgType:     InboundMessageTypeChat,
	}

	bus.PublishInbound(msg)

	if msg.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}

	select {
	case received := <-bus.inboundQueue:
		if received.ChannelName != "slack" {
			t.Errorf("expected ChannelName 'slack', got %s", received.ChannelName)
		}
	default:
		t.Errorf("expected message in queue")
	}
}

func TestMessageBusGetInbound(t *testing.T) {
	bus := NewMessageBus()

	msg := &InboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		UserID:      "U123",
		Text:        "hello",
	}

	bus.PublishInbound(msg)

	ctx := context.Background()
	received, err := bus.GetInbound(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if received.ChannelName != "slack" {
		t.Errorf("expected ChannelName 'slack', got %s", received.ChannelName)
	}
}

func TestMessageBusGetInboundTimeout(t *testing.T) {
	bus := NewMessageBus()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := bus.GetInbound(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestMessageBusSubscribeOutbound(t *testing.T) {
	bus := NewMessageBus()

	var received *OutboundMessage
	var mu sync.Mutex

	handler := func(msg *OutboundMessage) {
		mu.Lock()
		received = msg
		mu.Unlock()
	}

	bus.SubscribeOutbound(handler)

	msg := &OutboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		ThreadID:    "T123",
		Text:        "response",
	}

	bus.PublishOutbound(msg)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if received == nil {
		t.Errorf("expected message to be received")
	} else if received.ChannelName != "slack" {
		t.Errorf("expected ChannelName 'slack', got %s", received.ChannelName)
	}
	mu.Unlock()
}

func TestMessageBusMultipleHandlers(t *testing.T) {
	bus := NewMessageBus()

	var count int
	var mu sync.Mutex

	handler1 := func(msg *OutboundMessage) {
		mu.Lock()
		count++
		mu.Unlock()
	}
	handler2 := func(msg *OutboundMessage) {
		mu.Lock()
		count++
		mu.Unlock()
	}

	bus.SubscribeOutbound(handler1)
	bus.SubscribeOutbound(handler2)

	msg := &OutboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		Text:        "response",
	}

	bus.PublishOutbound(msg)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if count != 2 {
		t.Errorf("expected 2 handlers to be called, got %d", count)
	}
	mu.Unlock()
}
