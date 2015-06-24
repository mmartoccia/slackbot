package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Request(method string, url string, data url.Values, headers map[string]string) ([]byte, error) {
	var dataIn io.Reader

	if data == nil {
		dataIn = nil
	} else {
		dataIn = bytes.NewBufferString(data.Encode())
	}

	req, err := http.NewRequest(method, url, dataIn)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, err
}
