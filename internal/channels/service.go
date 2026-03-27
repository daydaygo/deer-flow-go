package channels

import (
	"context"
	"log"
	"sync"

	"github.com/user/deer-flow-go/internal/agent"
	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/store"
)

type Service struct {
	bus         *MessageBus
	store       *ChannelStore
	threadStore *store.ThreadStore
	engine      *agent.Engine
	manager     *Manager
	channels    map[string]Channel
	config      config.ChannelsConfig
	running     bool
	mu          sync.RWMutex
}

func NewService(cfg config.ChannelsConfig, threadStore *store.ThreadStore, engine *agent.Engine, dataDir string) *Service {
	storePath := ""
	if dataDir != "" {
		storePath = dataDir + "/channels/store.json"
	}
	return &Service{
		bus:         NewMessageBus(),
		store:       NewChannelStore(storePath),
		threadStore: threadStore,
		engine:      engine,
		channels:    make(map[string]Channel),
		config:      cfg,
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.manager = NewManager(s.bus, s.store, s.threadStore, s.engine)
	if err := s.manager.Start(ctx); err != nil {
		return err
	}

	for name, channelCfg := range s.config.Channels {
		if !channelCfg.Enabled {
			log.Printf("[Service] channel %s is disabled, skipping", name)
			continue
		}

		if err := s.startChannel(ctx, name, channelCfg); err != nil {
			log.Printf("[Service] failed to start channel %s: %v", name, err)
			continue
		}
	}

	s.running = true
	log.Printf("[Service] started with channels: %v", s.listChannelNames())
	return nil
}

func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	for name, channel := range s.channels {
		if err := channel.Stop(); err != nil {
			log.Printf("[Service] error stopping channel %s: %v", name, err)
		}
	}
	s.channels = make(map[string]Channel)

	s.manager.Stop()
	s.running = false
	log.Printf("[Service] stopped")
}

func (s *Service) startChannel(ctx context.Context, name string, cfg config.ChannelConfig) error {
	var channel Channel

	switch name {
	case "slack":
		slackCfg := SlackConfig{
			BotToken:     cfg.BotToken,
			AppToken:     cfg.AppToken,
			AllowedUsers: cfg.AllowedUsers,
		}
		channel = NewSlackChannel(s.bus, slackCfg)
	default:
		return ErrUnknownChannel
	}

	if err := channel.Start(ctx); err != nil {
		return err
	}

	s.channels[name] = channel
	log.Printf("[Service] channel %s started", name)
	return nil
}

func (s *Service) RestartChannel(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if channel, ok := s.channels[name]; ok {
		channel.Stop()
		delete(s.channels, name)
	}

	cfg, ok := s.config.Channels[name]
	if !ok {
		log.Printf("[Service] no config for channel %s", name)
		return ErrUnknownChannel
	}

	return s.startChannel(ctx, name, cfg)
}

func (s *Service) GetStatus() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channelsStatus := make(map[string]any)
	for name := range s.config.Channels {
		cfg := s.config.Channels[name]
		channel, exists := s.channels[name]
		running := exists && channel.IsRunning()
		channelsStatus[name] = map[string]any{
			"enabled": cfg.Enabled,
			"running": running,
		}
	}

	return map[string]any{
		"service_running": s.running,
		"channels":        channelsStatus,
	}
}

func (s *Service) listChannelNames() []string {
	names := make([]string, 0, len(s.channels))
	for name := range s.channels {
		names = append(names, name)
	}
	return names
}
