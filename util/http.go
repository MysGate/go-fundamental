package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var httpOnce sync.Once
var netClient *http.Client

func GetHTTPClient() *http.Client {
	httpOnce.Do(func() {
		var netTransport = &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		netClient = &http.Client{
			Timeout:   time.Second * 20,
			Transport: netTransport,
		}

	})

	return netClient
}

// HTTPReq ..
func HTTPReq(method string, url string, httpClient *http.Client, content []byte, headers map[string]string) (body []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(content))

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("HTTP get failed. err = %v, url = %s", err, url)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP read body failed. status code:%+v, url = %s", resp.StatusCode, url)
		return
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("HTTP read body failed. err = %v, url = %s", err, url)
		return
	}

	return
}
