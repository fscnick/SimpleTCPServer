package main

import "net/http"

type Context struct {
	buff		[]byte
	httpClient	*http.Client
}

