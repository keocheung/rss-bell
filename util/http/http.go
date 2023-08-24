package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Get(url string, headers map[string]string) ([]byte, error)
	Post(url string, body []byte, headers map[string]string) ([]byte, error)
}

func NewHTTPClient() HTTPClient {
	return &httpClient{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

type httpClient struct {
	client *http.Client
}

func (c *httpClient) Get(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %v", url, err)
	}
	for key, header := range headers {
		req.Header.Set(key, header)
	}
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %v", url, err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %v", url, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s error: %v %s", url, rsp.StatusCode, body)
	}
	return body, nil
}

func (c *httpClient) Post(url string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("post %s error: %v", url, err)
	}
	for key, header := range headers {
		req.Header.Set(key, header)
	}
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post %s error: %v", url, err)
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("post %s error: %v", url, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("post %s error: %v %s", url, rsp.StatusCode, rspBody)
	}
	return rspBody, nil
}
