package channels

import "errors"

var (
	ErrMissingToken   = errors.New("missing bot_token or app_token")
	ErrNotConnected   = errors.New("channel not connected")
	ErrUnknownChannel = errors.New("unknown channel type")
)
