package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type peerChannelsNotifySvc struct {
	cfg           *config.PeerChannels
	mu            sync.RWMutex
	pcSvc         payd.PeerChannelsService
	subscriptions map[string]*payd.PeerChannelSubscription
	handlers      map[payd.PeerChannelHandlerType]payd.PeerChannelsMessageHandler
}

// NewPeerChannelsNotifyService build and return a new peer channels notify service.
func NewPeerChannelsNotifyService(cfg *config.PeerChannels, pcSvc payd.PeerChannelsService) payd.PeerChannelsNotifyService {
	return &peerChannelsNotifySvc{
		cfg:           cfg,
		pcSvc:         pcSvc,
		subscriptions: make(map[string]*payd.PeerChannelSubscription),
		handlers:      make(map[payd.PeerChannelHandlerType]payd.PeerChannelsMessageHandler),
	}
}

func (p *peerChannelsNotifySvc) RegisterHandler(ht payd.PeerChannelHandlerType, hdlr payd.PeerChannelsMessageHandler) payd.PeerChannelsNotifyService {
	p.handlers[ht] = hdlr
	return p
}

func (p *peerChannelsNotifySvc) Subscribe(ctx context.Context, channel *payd.PeerChannel) error {
	if _, ok := p.handlers[channel.Type]; !ok {
		return fmt.Errorf("unrecognised channel type '%s'", string(channel.Type))
	}

	if channel.CreatedAt.IsZero() {
		channel.CreatedAt = time.Now()
	}

	log.Debug().Msgf("subscribing to channel %s with token %s at %s%s", channel.ID, channel.Token, channel.Host, channel.Path)

	u := url.URL{
		Scheme: "ws",
		Host:   channel.Host,
		Path:   path.Join(channel.Path, "/api/v1/channel", channel.ID, "/notify"),
	}
	q := u.Query()
	q.Set("token", channel.Token)
	u.RawQuery = q.Encode()

	lCtx, cancel := context.WithDeadline(context.Background(), channel.CreatedAt.Add(p.cfg.TTL))
	if _, ok := lCtx.Deadline(); !ok {
		log.Info().Msgf("deadline exceeded closing channel %s", channel.ID)
		defer cancel()
		return p.pcSvc.CloseChannel(ctx, channel.ID)
	}

	ws, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		defer cancel()
		if errors.Is(err, websocket.ErrBadHandshake) {
			return errors.Wrapf(err, "notification subscription handshake failed %d", resp.StatusCode)
		}
		return errors.Wrapf(err, "error dailing websocket")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	sub := payd.PeerChannelSubscription{
		Host:        channel.Host,
		Path:        channel.Path,
		ChannelID:   channel.ID,
		ChannelType: channel.Type,
		Token:       channel.Token,
		Conn:        ws,
	}

	func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.subscriptions[channel.ID] = &sub
	}()

	go p.listen(lCtx, &sub, cancel) //nolint:contextcheck // new context needed
	return nil
}

func (p *peerChannelsNotifySvc) listen(ctx context.Context, sub *payd.PeerChannelSubscription, cancel context.CancelFunc) {
	defer func() {
		_ = sub.Conn.Close()
	}()
	defer p.cleanup(sub.ChannelID)

	defer cancel()

	in := make(chan bool)

	go func() {
		if _, _, err := sub.Conn.ReadMessage(); err != nil {
			log.Error().Err(errors.WithStack(err))
		}

		log.Debug().Msgf("message received on channel %s", sub.ChannelID)
		in <- true
	}()

	for {
		select {
		case <-ctx.Done():
			log.Error().Err(p.pcSvc.CloseChannel(context.Background(), sub.ChannelID)) //nolint:contextcheck // new context needed
			return
		case <-in:
			log.Error().Err(errors.WithStack(p.handleNotification(context.Background(), sub, cancel))) //nolint:contextcheck // new context needed
		}
	}
}

func (p *peerChannelsNotifySvc) handleNotification(ctx context.Context, sub *payd.PeerChannelSubscription, cancel context.CancelFunc) error {
	msgs, err := p.pcSvc.PeerChannelsMessage(ctx, &payd.PeerChannelMessageArgs{
		ChannelID: sub.ChannelID,
		Host:      sub.Host,
		Path:      sub.Path,
		Token:     sub.Token,
	})
	if err != nil {
		log.Error().Err(errors.WithStack(err))
		return err
	}

	log.Debug().Msgf("channel %s fetched messages: %#v", sub.ChannelID, msgs)

	hdlr := p.handlers[sub.ChannelType]
	finished, err := hdlr.HandlePeerChannelsMessage(ctx, msgs)
	if err != nil {
		log.Error().Err(errors.WithStack(err))
		return err
	}

	if finished {
		defer cancel()
	}

	return nil
}

func (p *peerChannelsNotifySvc) cleanup(sub string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscriptions, sub)
}
