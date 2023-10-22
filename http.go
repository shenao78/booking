package booking

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func Get(url string, result interface{}) error {
	body, err := GetRaw(url)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func GetWithHeader(url string, header map[string]string, result interface{}) (http.Header, error) {
	return requestWithHeader("GET", url, header, nil, result)
}

func GetRaw(url string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func Post(url string, payload []byte, result interface{}) error {
	return PostWithHeader(url, nil, payload, result)
}

func PostWithHeader(url string, header map[string]string, payload []byte, result interface{}) error {
	_, err := requestWithHeader("POST", url, header, payload, result)
	return err
}

func requestWithHeader(method, url string, header map[string]string, payload []byte, result interface{}) (http.Header, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// set Content-Type in advance, and overwrite Content-Type if provided
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if result == nil {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return resp.Header, json.Unmarshal(body, result)
}
