//go:build darwin

package provider

import (
	"net"
)

func init() {
	net.DefaultResolver = &net.Resolver{PreferGo: false}
}
