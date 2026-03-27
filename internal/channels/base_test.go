package channels

import (
	"testing"
	"time"
)

func TestInboundMessageDefaults(t *testing.T) {
	msg := &InboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		UserID:      "U123",
		Text:        "hello",
	}

	if msg.MsgType != "" {
		t.Errorf("expected empty MsgType, got %s", msg.MsgType)
	}
}

func TestOutboundMessageDefaults(t *testing.T) {
	msg := &OutboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		ThreadID:    "T123",
		Text:        "response",
	}

	if msg.IsFinal != false {
		t.Errorf("expected IsFinal to be false by default")
	}
}

func TestInboundMessageTypeValues(t *testing.T) {
	if InboundMessageTypeChat != "chat" {
		t.Errorf("expected InboundMessageTypeChat to be 'chat'")
	}
	if InboundMessageTypeCommand != "command" {
		t.Errorf("expected InboundMessageTypeCommand to be 'command'")
	}
}

func TestMessageCreation(t *testing.T) {
	now := time.Now()
	msg := &InboundMessage{
		ChannelName: "slack",
		ChatID:      "C123",
		UserID:      "U123",
		Text:        "hello",
		MsgType:     InboundMessageTypeChat,
		ThreadTS:    "1234567890.123",
		TopicID:     "1234567890.123",
		CreatedAt:   now,
	}

	if msg.ChannelName != "slack" {
		t.Errorf("expected ChannelName 'slack', got %s", msg.ChannelName)
	}
	if msg.ChatID != "C123" {
		t.Errorf("expected ChatID 'C123', got %s", msg.ChatID)
	}
	if msg.ThreadTS != "1234567890.123" {
		t.Errorf("expected ThreadTS '1234567890.123', got %s", msg.ThreadTS)
	}
}
