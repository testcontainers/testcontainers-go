package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const dockerResponseLimit = 1 << 20

type dockerEngine struct {
	base   *url.URL
	client *http.Client
}

type dockerContainer struct {
	ID      string
	Name    string
	Labels  map[string]string
	Running bool
	Ports   []dockerPort
}

type dockerPort struct {
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func newDockerEngine(base *url.URL, client *http.Client) *dockerEngine {
	copy := *base
	return &dockerEngine{base: &copy, client: client}
}

func (e *dockerEngine) endpoint(path string, query url.Values) string {
	target := *e.base
	target.Path = path
	target.RawPath = ""
	target.RawQuery = query.Encode()
	return target.String()
}

func (e *dockerEngine) listByLease(ctx context.Context, token string) ([]dockerContainer, error) {
	filters, err := json.Marshal(map[string][]string{"label": {portLeaseLabel + "=" + token}})
	if err != nil {
		return nil, err
	}
	query := url.Values{"all": {"1"}, "filters": {string(filters)}}
	var rows []struct {
		ID     string            `json:"Id"`
		Names  []string          `json:"Names"`
		Labels map[string]string `json:"Labels"`
		Ports  []dockerPort      `json:"Ports"`
	}
	status, err := e.getJSON(ctx, "/containers/json", query, &rows)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("docker list returned HTTP %d", status)
	}
	containers := make([]dockerContainer, 0, len(rows))
	for _, row := range rows {
		name := ""
		if len(row.Names) > 0 {
			name = row.Names[0]
		}
		containers = append(containers, dockerContainer{ID: row.ID, Name: name, Labels: row.Labels, Ports: row.Ports})
	}
	return containers, nil
}

func (e *dockerEngine) knownPorts(ctx context.Context) (map[int]bool, error) {
	query := url.Values{"all": {"1"}}
	var rows []struct {
		Ports []dockerPort `json:"Ports"`
	}
	status, err := e.getJSON(ctx, "/containers/json", query, &rows)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("docker list returned HTTP %d", status)
	}
	ports := map[int]bool{2375: true}
	for _, row := range rows {
		for _, port := range row.Ports {
			if port.PrivatePort > 0 {
				ports[port.PrivatePort] = true
			}
			if port.PublicPort > 0 {
				ports[port.PublicPort] = true
			}
		}
	}
	return ports, nil
}

func (e *dockerEngine) inspect(ctx context.Context, ref string) (*dockerContainer, error) {
	path := "/containers/" + url.PathEscape(ref) + "/json"
	var row struct {
		ID     string `json:"Id"`
		Name   string `json:"Name"`
		Config struct {
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		State struct {
			Running bool   `json:"Running"`
			Status  string `json:"Status"`
		} `json:"State"`
	}
	status, err := e.getJSON(ctx, path, nil, &row)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, nil
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("docker inspect %q returned HTTP %d", ref, status)
	}
	return &dockerContainer{ID: row.ID, Name: row.Name, Labels: row.Config.Labels, Running: row.State.Running}, nil
}

func (e *dockerEngine) getJSON(ctx context.Context, path string, query url.Values, destination any) (int, error) {
	requestCtx, cancel := boundedContext(ctx, portOperationTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, e.endpoint(path, query), nil)
	if err != nil {
		return 0, err
	}
	response, err := e.client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, dockerResponseLimit))
		return response.StatusCode, nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, dockerResponseLimit))
		return response.StatusCode, errors.Join(fmt.Errorf("docker API HTTP %d: %s", response.StatusCode, body), readErr)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, dockerResponseLimit)).Decode(destination); err != nil {
		return response.StatusCode, fmt.Errorf("decode Docker API response (HTTP %s): %w", strconv.Itoa(response.StatusCode), err)
	}
	return response.StatusCode, nil
}
