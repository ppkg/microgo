package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

var (
	HttpDefaultClient = http.DefaultClient
	HttpClient        = &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)
