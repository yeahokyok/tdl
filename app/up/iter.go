package up

import (
	"context"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/uploader"
	"github.com/iyear/tdl/pkg/utils"
)

type file struct {
	file  string
	thumb string
}

type iter struct {
	files  []*file
	to     peers.Peer
	photo  bool
	remove bool

	cur  int
	err  error
	file uploader.Elem
}

func newIter(files []*file, to peers.Peer, photo, remove bool) *iter {
	return &iter{
		files:  files,
		to:     to,
		photo:  photo,
		remove: remove,

		cur:  0,
		err:  nil,
		file: nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.cur >= len(i.files) || i.err != nil {
		return false
	}

	cur := i.files[i.cur]
	i.cur++

	f, err := os.Open(cur.file)
	if err != nil {
		i.err = errors.Wrap(err, "open file")
		return false
	}

	var thumb *uploaderFile = nil
	// has thumbnail
	if cur.thumb != "" {
		tMime, err := mimetype.DetectFile(cur.thumb)
		if err != nil || !utils.Media.IsImage(tMime.String()) { // TODO(iyear): jpg only
			i.err = errors.Wrapf(err, "invalid thumbnail file: %v", cur.thumb)
			return false
		}
		thumbFile, err := os.Open(cur.thumb)
		if err != nil {
			i.err = errors.Wrap(err, "open thumbnail file")
			return false
		}

		thumb = &uploaderFile{File: thumbFile, size: 0}
	}

	stat, err := f.Stat()
	if err != nil {
		i.err = errors.Wrap(err, "stat file")
		return false
	}

	i.file = &iterElem{
		file:  &uploaderFile{File: f, size: stat.Size()},
		thumb: thumb,
		to:    i.to,

		asPhoto: i.photo,
		remove:  i.remove,
	}

	return true
}

func (i *iter) Value() uploader.Elem {
	return i.file
}

func (i *iter) Err() error {
	return i.err
}
