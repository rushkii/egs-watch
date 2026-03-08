package pkg

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	client *http.Client
}

func NewClient() *HttpClient {
	return &HttpClient{client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *HttpClient) GetJSON(url string, decode interface{}) error {
	r, err := c.client.Get(url)

	if (err) != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(decode)
}

func (c *HttpClient) Download(url string) ([]byte, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
