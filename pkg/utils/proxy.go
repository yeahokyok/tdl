package utils

import (
	"net/url"

	"github.com/go-faster/errors"
	"github.com/iyear/connectproxy"
	"golang.org/x/net/proxy"
)

type _proxy struct{}

var Proxy = _proxy{}

func init() {
	connectproxy.Register(&connectproxy.Config{
		InsecureSkipVerify: true,
	})
}

func (p _proxy) GetDial(_url string) (proxy.ContextDialer, error) {
	u, err := url.Parse(_url)
	if err != nil {
		return nil, errors.Wrap(err, "parse proxy url")
	}
	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, errors.Wrap(err, "proxy from url")
	}

	if d, ok := dialer.(proxy.ContextDialer); ok {
		return d, nil
	}

	return nil, errors.New("proxy dialer is not ContextDialer")
}
