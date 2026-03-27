package channels

import (
	"context"
	"time"
)

type InboundMessageType string

const (
	InboundMessageTypeChat    InboundMessageType = "chat"
	InboundMessageTypeCommand InboundMessageType = "command"
)

type InboundMessage struct {
	ChannelName string             `json:"channel_name"`
	ChatID      string             `json:"chat_id"`
	UserID      string             `json:"user_id"`
	Text        string             `json:"text"`
	MsgType     InboundMessageType `json:"msg_type"`
	ThreadTS    string             `json:"thread_ts,omitempty"`
	TopicID     string             `json:"topic_id,omitempty"`
	Files       []map[string]any   `json:"files,omitempty"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

type OutboundMessage struct {
	ChannelName string         `json:"channel_name"`
	ChatID      string         `json:"chat_id"`
	ThreadID    string         `json:"thread_id"`
	Text        string         `json:"text"`
	IsFinal     bool           `json:"is_final"`
	ThreadTS    string         `json:"thread_ts,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
	Send(ctx context.Context, msg *OutboundMessage) error
	IsRunning() bool
}
