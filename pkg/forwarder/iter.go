package forwarder

import (
	"context"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

type Iter interface {
	Next(ctx context.Context) bool
	Value() Elem
	Err() error
}

type Elem interface {
	Mode() Mode

	From() peers.Peer
	Msg() *tg.Message
	To() peers.Peer

	AsSilent() bool
	AsDryRun() bool
}
