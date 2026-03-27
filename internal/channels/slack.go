package channels

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type SlackChannel struct {
	name         string
	bus          *MessageBus
	config       SlackConfig
	running      bool
	socketClient *socketmode.Client
	webClient    *slack.Client
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	allowedUsers map[string]bool
}

type SlackConfig struct {
	BotToken     string   `mapstructure:"bot_token"`
	AppToken     string   `mapstructure:"app_token"`
	AllowedUsers []string `mapstructure:"allowed_users"`
}

func NewSlackChannel(bus *MessageBus, config SlackConfig) *SlackChannel {
	allowed := make(map[string]bool)
	for _, u := range config.AllowedUsers {
		allowed[u] = true
	}
	return &SlackChannel{
		name:         "slack",
		bus:          bus,
		config:       config,
		allowedUsers: allowed,
	}
}

func (c *SlackChannel) Name() string {
	return c.name
}

func (c *SlackChannel) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

func (c *SlackChannel) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	if c.config.BotToken == "" || c.config.AppToken == "" {
		return ErrMissingToken
	}

	c.webClient = slack.New(
		c.config.BotToken,
		slack.OptionAppLevelToken(c.config.AppToken),
	)
	c.socketClient = socketmode.New(c.webClient)

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.running = true
	c.bus.SubscribeOutbound(c.onOutbound)

	go c.runSocketMode()
	log.Printf("[Slack] channel started")
	return nil
}

func (c *SlackChannel) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	c.bus.UnsubscribeOutbound(c.onOutbound)
	if c.cancel != nil {
		c.cancel()
	}
	log.Printf("[Slack] channel stopped")
	return nil
}

func (c *SlackChannel) Send(ctx context.Context, msg *OutboundMessage) error {
	if c.webClient == nil {
		return ErrNotConnected
	}

	opts := []slack.MsgOption{slack.MsgOptionText(msg.Text, false)}
	if msg.ThreadTS != "" {
		opts = append(opts, slack.MsgOptionTS(msg.ThreadTS))
	}

	_, _, err := c.webClient.PostMessageContext(ctx, msg.ChatID, opts...)
	if err != nil {
		log.Printf("[Slack] failed to send message: %v", err)
		return err
	}

	if msg.ThreadTS != "" {
		c.addReaction(msg.ChatID, msg.ThreadTS, "white_check_mark")
	}
	return nil
}

func (c *SlackChannel) runSocketMode() {
	go c.socketClient.Run()

	for {
		select {
		case <-c.ctx.Done():
			return
		case evt, ok := <-c.socketClient.Events:
			if !ok {
				return
			}
			c.handleSocketEvent(evt)
		}
	}
}

func (c *SlackChannel) handleSocketEvent(evt socketmode.Event) {
	if evt.Type != socketmode.EventTypeEventsAPI {
		c.socketClient.Ack(*evt.Request, nil)
		return
	}

	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		c.socketClient.Ack(*evt.Request, nil)
		return
	}

	c.socketClient.Ack(*evt.Request, nil)

	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			c.handleMessageEvent(ev)
		case *slackevents.AppMentionEvent:
			c.handleAppMentionEvent(ev)
		}
	}
}

func (c *SlackChannel) handleMessageEvent(ev *slackevents.MessageEvent) {
	if ev.BotID != "" || ev.SubType != "" {
		return
	}

	if len(c.allowedUsers) > 0 && !c.allowedUsers[ev.User] {
		return
	}

	text := ev.Text
	if text == "" {
		return
	}

	msgType := InboundMessageTypeChat
	if len(text) > 0 && text[0] == '/' {
		msgType = InboundMessageTypeCommand
	}

	threadTS := ev.ThreadTimeStamp
	if threadTS == "" {
		threadTS = ev.TimeStamp
	}

	msg := &InboundMessage{
		ChannelName: c.name,
		ChatID:      ev.Channel,
		UserID:      ev.User,
		Text:        text,
		MsgType:     msgType,
		ThreadTS:    threadTS,
		TopicID:     threadTS,
	}

	c.addReaction(ev.Channel, ev.TimeStamp, "eyes")
	c.sendRunningReply(ev.Channel, threadTS)
	c.bus.PublishInbound(msg)
}

func (c *SlackChannel) handleAppMentionEvent(ev *slackevents.AppMentionEvent) {
	if len(c.allowedUsers) > 0 && !c.allowedUsers[ev.User] {
		return
	}

	text := ev.Text
	if text == "" {
		return
	}

	msgType := InboundMessageTypeChat
	if len(text) > 0 && text[0] == '/' {
		msgType = InboundMessageTypeCommand
	}

	threadTS := ev.ThreadTimeStamp
	if threadTS == "" {
		threadTS = ev.TimeStamp
	}

	msg := &InboundMessage{
		ChannelName: c.name,
		ChatID:      ev.Channel,
		UserID:      ev.User,
		Text:        text,
		MsgType:     msgType,
		ThreadTS:    threadTS,
		TopicID:     threadTS,
	}

	c.addReaction(ev.Channel, ev.TimeStamp, "eyes")
	c.sendRunningReply(ev.Channel, threadTS)
	c.bus.PublishInbound(msg)
}

func (c *SlackChannel) onOutbound(msg *OutboundMessage) {
	if msg.ChannelName != c.name {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.Send(ctx, msg); err != nil {
		log.Printf("[Slack] failed to send outbound message: %v", err)
	}
}

func (c *SlackChannel) addReaction(channel, timestamp, emoji string) {
	if c.webClient == nil {
		return
	}
	err := c.webClient.AddReaction(emoji, slack.ItemRef{Channel: channel, Timestamp: timestamp})
	if err != nil && !isAlreadyReactedError(err) {
		log.Printf("[Slack] failed to add reaction: %v", err)
	}
}

func (c *SlackChannel) sendRunningReply(channel, threadTS string) {
	if c.webClient == nil {
		return
	}
	opts := []slack.MsgOption{
		slack.MsgOptionText(":hourglass_flowing_sand: Working on it...", false),
		slack.MsgOptionTS(threadTS),
	}
	_, _, err := c.webClient.PostMessage(channel, opts...)
	if err != nil {
		log.Printf("[Slack] failed to send running reply: %v", err)
	}
}

func isAlreadyReactedError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "already_reacted" || err.Error() == "invalid_reaction"
}
