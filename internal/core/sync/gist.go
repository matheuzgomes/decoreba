package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const gistAPI = "https://api.github.com/gists"

type GistBackend struct {
	token string
}

func NewGistBackend(token string) *GistBackend {
	return &GistBackend{token: token}
}

func (g *GistBackend) Name() string { return "gist" }

func (g *GistBackend) Upload(data []byte, remoteID string) (string, error) {
	if remoteID == "" {
		return g.create(data)
	}
	return remoteID, g.update(remoteID, data)
}

func (g *GistBackend) Download(remoteID string) ([]byte, error) {
	url := gistAPI + "/" + remoteID
	gist, err := g.do("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	files, ok := gist["files"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("download: no files in gist")
	}
	file, ok := files["commands.json"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("download: commands.json not found in gist")
	}
	content, ok := file["content"].(string)
	if !ok {
		return nil, fmt.Errorf("download: commands.json has no content")
	}
	return []byte(content), nil
}

func (g *GistBackend) Delete(remoteID string) error {
	url := gistAPI + "/" + remoteID
	_, err := g.do("DELETE", url, nil)
	return err
}

func (g *GistBackend) create(data []byte) (string, error) {
	body := map[string]interface{}{
		"description": "decoreba command vault",
		"public":      false,
		"files": map[string]interface{}{
			"commands.json": map[string]string{
				"content": string(data),
			},
		},
	}
	gist, err := g.do("POST", gistAPI, body)
	if err != nil {
		return "", fmt.Errorf("create gist: %w", err)
	}
	id, _ := gist["id"].(string)
	if id == "" {
		return "", fmt.Errorf("create gist: no id in response")
	}
	return id, nil
}

func (g *GistBackend) update(remoteID string, data []byte) error {
	url := gistAPI + "/" + remoteID
	body := map[string]interface{}{
		"files": map[string]interface{}{
			"commands.json": map[string]string{
				"content": string(data),
			},
		},
	}
	_, err := g.do("PATCH", url, body)
	if err != nil {
		return fmt.Errorf("update gist: %w", err)
	}
	return nil
}

func (g *GistBackend) do(method, url string, body interface{}) (map[string]interface{}, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = strings.NewReader(string(b))
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "decoreba")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if method == "DELETE" {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}
