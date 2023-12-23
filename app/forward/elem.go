package forward

import (
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/pkg/forwarder"
)

type iterElem struct {
	from peers.Peer
	msg  *tg.Message
	to   peers.Peer

	opts iterOptions
}

func (i *iterElem) Mode() forwarder.Mode { return i.opts.mode }

func (i *iterElem) From() peers.Peer { return i.from }

func (i *iterElem) Msg() *tg.Message { return i.msg }

func (i *iterElem) To() peers.Peer { return i.to }

func (i *iterElem) AsSilent() bool { return i.opts.silent }

func (i *iterElem) AsDryRun() bool { return i.opts.dryRun }
