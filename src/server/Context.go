package main

import "net/http"
import "net"

type Context struct {
	buff		[]byte
	conn		net.Conn
	httpClient	*http.Client
	dstUrl		string
}

