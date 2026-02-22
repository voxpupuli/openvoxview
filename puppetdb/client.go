package puppetdb

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
)

type client struct {
}

type PdbQuery struct {
	Query       []any  `json:"query"`
	SummarizeBy string `json:"summarize_by,omitempty"`
}

type PdbBadQueryError error

func NewClient() *client {
	return &client{}
}

func (c *client) call(httpMethod string, endpoint string, payload any, query url.Values, responseData any) (*http.Response, int, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	uri := fmt.Sprintf("%s/%s", cfg.GetPuppetDbAddress(), endpoint)
	if query != nil {
		uri = fmt.Sprintf("%s?%s", uri, query.Encode())
	}

	var data []byte

	if payload != nil {
		data, err = json.Marshal(&payload)
		if err != nil {
			fmt.Printf("err: %s", err)
		}
		fmt.Printf("Payload:\n%s\n", data)
	}

	fmt.Printf("HTTP: %#v: %#v\n", httpMethod, uri)

	var tlsConfig *tls.Config

	if cfg.PuppetDB.TLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.PuppetDB.TLSIgnore,
		}

		if cfg.PuppetDB.TLS_CA != "" {
			caCert, err := os.ReadFile(cfg.PuppetDB.TLS_CA)
			if err != nil {
				return nil, 0, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		if cfg.PuppetDB.TLS_KEY != "" {
			cer, err := tls.LoadX509KeyPair(cfg.PuppetDB.TLS_CERT, cfg.PuppetDB.TLS_KEY)
			if err != nil {
				return nil, 0, err
			}

			tlsConfig.Certificates = []tls.Certificate{cer}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest(httpMethod, uri, bytes.NewBuffer(data))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return resp, http.StatusInternalServerError, err
	}

	defer resp.Body.Close()

	responseRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, resp.StatusCode, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		err = json.Unmarshal(responseRaw, responseData)
		if err != nil {
			return resp, resp.StatusCode, err
		}
	case http.StatusBadRequest:
		return nil, resp.StatusCode, errors.New(string(responseRaw))
	}
	return resp, resp.StatusCode, nil
}

func (c *client) Query(query string) ([]json.RawMessage, int, error) {
	type PuppetDbQueryRequest struct {
		Query string `json:"query"`
	}

	requestBody := PuppetDbQueryRequest{
		Query: query,
	}

	resp := []json.RawMessage{}

	_, code, err := c.call(http.MethodPost, "pdb/query/v4", &requestBody, nil, &resp)

	return resp, code, err
}

func (c *client) GetFacts(query *PdbQuery) ([]model.Fact, error) {
	var resp []model.Fact
	_, _, err := c.call(http.MethodPost, "pdb/query/v4/facts", query, nil, &resp)
	return resp, err
}

func (c *client) GetFactNames() (json.RawMessage, error) {
	resp := json.RawMessage{}
	_, _, err := c.call(http.MethodGet, "pdb/query/v4/fact-names", nil, nil, &resp)
	return resp, err
}

func (c *client) GetEventCounts(query *PdbQuery) ([]model.EventCount, error) {
	var resp []model.EventCount
	_, _, err := c.call(http.MethodPost, "pdb/query/v4/event-counts", query, nil, &resp)
	return resp, err
}

func (c *client) GetNodes(query *PdbQuery) ([]model.Node, error) {
	var resp []model.Node
	_, _, err := c.call(http.MethodPost, "pdb/query/v4/nodes", query, nil, &resp)
	return resp, err
}

func (c *client) GetMetric(metricName string) (model.Metric, error) {
	var resp model.Metric
	_, _, err := c.call(http.MethodGet, fmt.Sprintf("metrics/v2/%s", metricName), nil, nil, &resp)
	return resp, err
}

func (c *client) GetMetricList() (model.MetricList, error) {
	var resp model.MetricList
	_, _, err := c.call(http.MethodGet, "metrics/v2/list", nil, nil, &resp)
	return resp, err
}

func (c *client) DeactivateNode(certname string) (*model.CommandResponse, error) {
	payload := model.DeactivateNodePayload{
		Certname:          certname,
		ProducerTimestamp: time.Now().UTC(),
	}

	query := url.Values{}
	query.Set("command", "deactivate node")
	query.Set("version", "3")
	query.Set("certname", certname)

	var resp model.CommandResponse

	_, statusCode, err := c.call(http.MethodPost, "pdb/cmd/v1", payload, query, &resp)

	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}

	return &resp, nil
}
