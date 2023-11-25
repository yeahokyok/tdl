package dliter

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/tplfunc"
	"github.com/iyear/tdl/pkg/utils"
)

func New(ctx context.Context, opts *Options) (*Iter, error) {
	tpl, err := template.New("dl").
		Funcs(tplfunc.FuncMap(tplfunc.All...)).
		Parse(opts.Template)
	if err != nil {
		return nil, err
	}

	dialogs := collectDialogs(opts.Dialogs)
	// if msgs is empty, return error to avoid range out of index
	if len(dialogs) == 0 {
		return nil, fmt.Errorf("you must specify at least one message")
	}

	// include and exclude
	includeMap := filterMap(opts.Include, utils.FS.AddPrefixDot)
	excludeMap := filterMap(opts.Exclude, utils.FS.AddPrefixDot)

	// to keep fingerprint stable
	sortDialogs(dialogs, opts.Desc)

	manager := peers.Options{Storage: storage.NewPeers(opts.KV)}.Build(opts.Pool.Default(ctx))
	it := &Iter{
		pool:        opts.Pool,
		dialogs:     dialogs,
		include:     includeMap,
		exclude:     excludeMap,
		curi:        0,
		curj:        -1,
		preSum:      preSum(dialogs),
		finished:    make(map[int]struct{}),
		template:    tpl,
		manager:     manager,
		fingerprint: fingerprint(dialogs),
	}

	return it, nil
}

func (iter *Iter) Next(ctx context.Context) (*downloader.Item, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	iter.mu.Lock()
	iter.curj++
	if iter.curj >= len(iter.dialogs[iter.curi].Messages) {
		if iter.curi++; iter.curi >= len(iter.dialogs) {
			return nil, errors.New("no more items")
		}
		iter.curj = 0
	}
	i, j := iter.curi, iter.curj
	iter.mu.Unlock()

	// check if finished
	if _, ok := iter.finished[iter.ij2n(i, j)]; ok {
		return nil, downloader.ErrSkip
	}

	return iter.item(ctx, i, j)
}

func (iter *Iter) item(ctx context.Context, i, j int) (*downloader.Item, error) {
	peer, msg := iter.dialogs[i].Peer, iter.dialogs[i].Messages[j]

	id := utils.Telegram.GetInputPeerID(peer)

	message, err := utils.Telegram.GetSingleMessage(ctx, iter.pool.Default(ctx), peer, msg)
	if err != nil {
		return nil, errors.Wrap(err, "resolve message")
	}

	item, ok := tmedia.GetMedia(message)
	if !ok {
		return nil, fmt.Errorf("can not get media from %d/%d message",
			id, message.ID)
	}

	// process include and exclude
	ext := filepath.Ext(item.Name)
	if len(iter.include) > 0 {
		if _, ok = iter.include[ext]; !ok {
			return nil, downloader.ErrSkip
		}
	}
	if len(iter.exclude) > 0 {
		if _, ok = iter.exclude[ext]; ok {
			return nil, downloader.ErrSkip
		}
	}

	buf := bytes.Buffer{}

	authorPeer, _ := message.GetFromID()
	AuthorPeerID := utils.Telegram.GetPeerID(authorPeer)
	err = iter.template.Execute(&buf, &fileTemplate{
		DialogID:     id,
		MessageID:    message.ID,
		MessageDate:  int64(message.Date),
		FileName:     item.Name,
		FileSize:     utils.Byte.FormatBinaryBytes(item.Size),
		DownloadDate: time.Now().Unix(),
		AuthorID:     AuthorPeerID,
	})
	if err != nil {
		return nil, err
	}
	item.Name = buf.String()

	return &downloader.Item{
		ID:    iter.ij2n(i, j),
		Media: item,
	}, nil
}

func (iter *Iter) Finish(_ context.Context, id int) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished[id] = struct{}{}
	return nil
}

func (iter *Iter) Total(_ context.Context) int {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	total := 0
	for _, m := range iter.dialogs {
		total += len(m.Messages)
	}
	return total
}

func (iter *Iter) ij2n(i, j int) int {
	return iter.preSum[i] + j
}

func (iter *Iter) SetFinished(finished map[int]struct{}) {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	iter.finished = finished
}

func (iter *Iter) Finished() map[int]struct{} {
	iter.mu.Lock()
	defer iter.mu.Unlock()

	return iter.finished
}

func (iter *Iter) Fingerprint() string {
	return iter.fingerprint
}
