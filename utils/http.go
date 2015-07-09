package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Request(method string, reqUrl string, data url.Values, headers map[string]string) ([]byte, error) {
	var dataIn io.Reader

	if data == nil {
		dataIn = nil
	} else {
		// s, err := url.QueryUnescape(data.Encode())
		// if err != nil {
		// 	return nil, err
		// }
		// dataIn = bytes.NewBufferString(s)
		dataIn = bytes.NewBufferString(data.Encode())
	}

	return RequestRaw(method, reqUrl, dataIn, headers)
}

func RequestRaw(method string, url string, dataIn io.Reader, headers map[string]string) ([]byte, error) {
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
