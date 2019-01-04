package infrastructure

import (
	"fmt"

	"golang.org/x/net/proxy"
)

// ProxyDialerHandler struct to implements proxy dialer operations
type ProxyDialerHandler struct{}

// NewProxyDialerHandler create a instance of a golang/net proxy dialer
func NewProxyDialerHandler(connType, host string) (interface{}, error) {
	dialer, err := proxy.SOCKS5(connType, host, nil, proxy.Direct)
	if err != nil {
		err = fmt.Errorf("Was not possible connect with proxy: %+v. Error: %+v",
			host, err)
	}
	return dialer, err
}
