package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func TestRequest(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	io.Copy(io.Discard, resp.Body)
	return nil
}
