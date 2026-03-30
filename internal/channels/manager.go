package channels

import (
	"context"
	"log"
	"sync"

	"github.com/user/deer-flow-go/internal/agent"
	"github.com/user/deer-flow-go/internal/store"
)

type Manager struct {
	bus         *MessageBus
	store       *ChannelStore
	threadStore *store.ThreadStore
	engine      *agent.Engine
	running     bool
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

func NewManager(bus *MessageBus, store *ChannelStore, threadStore *store.ThreadStore, engine *agent.Engine) *Manager {
	return &Manager{
		bus:         bus,
		store:       store,
		threadStore: threadStore,
		engine:      engine,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true
	m.wg.Add(1)
	go m.dispatchLoop(ctx)
	log.Printf("[Manager] started")
	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
	m.wg.Wait()
	log.Printf("[Manager] stopped")
}

func (m *Manager) dispatchLoop(ctx context.Context) {
	defer m.wg.Done()
	log.Printf("[Manager] dispatch loop started")

	for {
		m.mu.RLock()
		running := m.running
		m.mu.RUnlock()

		if !running {
			return
		}

		msg, err := m.bus.GetInbound(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			log.Printf("[Manager] error getting inbound message: %v", err)
			continue
		}

		log.Printf("[Manager] received inbound: channel=%s, chat_id=%s, type=%s", msg.ChannelName, msg.ChatID, msg.MsgType)

		m.wg.Go(func() {
			m.handleMessage(ctx, msg)
		})
	}
}

func (m *Manager) handleMessage(ctx context.Context, msg *InboundMessage) {
	if msg.MsgType == InboundMessageTypeCommand {
		m.handleCommand(ctx, msg)
	} else {
		m.handleChat(ctx, msg)
	}
}

func (m *Manager) handleChat(ctx context.Context, msg *InboundMessage) {
	threadID := m.store.GetThreadID(msg.ChannelName, msg.ChatID, msg.TopicID)

	if threadID == "" {
		thread, err := m.threadStore.Create()
		if err != nil {
			log.Printf("[Manager] failed to create thread: %v", err)
			m.sendError(msg, "Failed to create conversation")
			return
		}
		threadID = thread.ThreadID
		m.store.SetThreadID(msg.ChannelName, msg.ChatID, threadID, msg.TopicID, msg.UserID)
		log.Printf("[Manager] created new thread: thread_id=%s", threadID)
	} else {
		log.Printf("[Manager] reusing thread: thread_id=%s", threadID)
	}

	response, err := m.engine.Invoke(ctx, threadID, msg.Text)
	if err != nil {
		log.Printf("[Manager] failed to invoke engine: %v", err)
		m.sendError(msg, "An internal error occurred")
		return
	}

	responseText := response.Content
	if responseText == "" {
		responseText = "(No response from agent)"
	}

	outbound := &OutboundMessage{
		ChannelName: msg.ChannelName,
		ChatID:      msg.ChatID,
		ThreadID:    threadID,
		Text:        responseText,
		IsFinal:     true,
		ThreadTS:    msg.ThreadTS,
	}

	m.bus.PublishOutbound(outbound)
}

func (m *Manager) handleCommand(ctx context.Context, msg *InboundMessage) {
	text := msg.Text
	command := parseCommand(text)

	var reply string
	switch command {
	case "new":
		thread, err := m.threadStore.Create()
		if err != nil {
			reply = "Failed to create new conversation"
		} else {
			m.store.SetThreadID(msg.ChannelName, msg.ChatID, thread.ThreadID, msg.TopicID, msg.UserID)
			reply = "New conversation started."
			log.Printf("[Manager] new thread created: thread_id=%s", thread.ThreadID)
		}
	case "status":
		threadID := m.store.GetThreadID(msg.ChannelName, msg.ChatID, msg.TopicID)
		if threadID != "" {
			reply = "Active thread: " + threadID
		} else {
			reply = "No active conversation."
		}
	case "help":
		reply = "Available commands:\n/new — Start a new conversation\n/status — Show current thread info\n/help — Show this help"
	default:
		reply = "Unknown command: /" + command + ". Type /help for available commands."
	}

	outbound := &OutboundMessage{
		ChannelName: msg.ChannelName,
		ChatID:      msg.ChatID,
		ThreadID:    m.store.GetThreadID(msg.ChannelName, msg.ChatID, msg.TopicID),
		Text:        reply,
		IsFinal:     true,
		ThreadTS:    msg.ThreadTS,
	}

	m.bus.PublishOutbound(outbound)
}

func (m *Manager) sendError(msg *InboundMessage, errorText string) {
	outbound := &OutboundMessage{
		ChannelName: msg.ChannelName,
		ChatID:      msg.ChatID,
		ThreadID:    m.store.GetThreadID(msg.ChannelName, msg.ChatID, msg.TopicID),
		Text:        errorText,
		IsFinal:     true,
		ThreadTS:    msg.ThreadTS,
	}
	m.bus.PublishOutbound(outbound)
}

func parseCommand(text string) string {
	if len(text) == 0 || text[0] != '/' {
		return ""
	}
	cmd := text[1:]
	for i, c := range cmd {
		if c == ' ' {
			return cmd[:i]
		}
	}
	return cmd
}
