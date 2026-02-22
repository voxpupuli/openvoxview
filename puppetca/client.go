package puppetca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
)

type Client struct {
	config    *config.Config
	transport *http.Transport
}

func NewClient(config *config.Config) *Client {
	return &Client{
		config:    config,
		transport: nil,
	}
}

func (c *Client) call(httpMethod string, endpoint string, payload any, query url.Values, responseData any) (*http.Response, int, error) {
	uri := fmt.Sprintf("%s/%s", c.config.GetPuppetCAAddress(), endpoint)
	if query != nil {
		uri = fmt.Sprintf("%s?%s", uri, query.Encode())
	}

	var data []byte
	var err error

	if payload != nil {
		data, err = json.Marshal(&payload)
		if err != nil {
			fmt.Printf("err: %s", err)
		}
		fmt.Printf("Payload:\n%s\n", data)
	}

	fmt.Printf("HTTP: %#v: %#v\n", httpMethod, uri)

	if c.transport == nil {
		var tlsConfig *tls.Config

		if c.config.PuppetCA.TLS {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: c.config.PuppetCA.TLSIgnore,
			}

			if c.config.PuppetCA.TLS_CA != "" {
				caCert, err := os.ReadFile(c.config.PuppetCA.TLS_CA)
				if err != nil {
					return nil, 0, err
				}
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				tlsConfig.RootCAs = caCertPool
			}

			if c.config.PuppetCA.TLS_KEY != "" {
				cer, err := tls.LoadX509KeyPair(c.config.PuppetCA.TLS_CERT, c.config.PuppetCA.TLS_KEY)
				if err != nil {
					return nil, 0, err
				}

				tlsConfig.Certificates = []tls.Certificate{cer}
			}
		}

		c.transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	httpClient := &http.Client{
		Transport: c.transport,
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
		if responseData != nil {
			err = json.Unmarshal(responseRaw, responseData)
			if err != nil {
				return resp, resp.StatusCode, err
			}
		}
	}
	return resp, resp.StatusCode, nil
}

func (c *Client) GetCertificates(state *model.CertificateState) ([]model.CertificateStatus, error) {
	var resp []model.CertificateStatus

	query := url.Values{}

	if state != nil {
		query.Set("state", state.String())
	}

	_, _, err := c.call(http.MethodGet, "puppet-ca/v1/certificate_statuses/all", nil, query, &resp)

	return resp, err
}

func (c *Client) GetCertificate(name string) (*model.CertificateStatus, error) {
	var resp model.CertificateStatus

	_, statusCode, err := c.call(http.MethodGet, fmt.Sprintf("puppet-ca/v1/certificate_status/%s", name), nil, nil, &resp)

	switch statusCode {
	case http.StatusOK:
		return &resp, err
	default:
	}

	return nil, err
}

func (c *Client) SignCertificate(name string) error {
	payload := struct {
		DesiredState string `json:"desired_state"`
	}{
		DesiredState: "signed",
	}

	_, statusCode, err := c.call(http.MethodPut, fmt.Sprintf("puppet-ca/v1/certificate_status/%s", name), payload, nil, nil)

	if err != nil {
		return err
	}

	switch statusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}

func (c *Client) RevokeCertificate(name string) error {
	payload := struct {
		DesiredState string `json:"desired_state"`
	}{
		DesiredState: "revoked",
	}

	_, statusCode, err := c.call(http.MethodPut, fmt.Sprintf("puppet-ca/v1/certificate_status/%s", name), payload, nil, nil)

	if err != nil {
		return err
	}

	switch statusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}

func (c *Client) CleanCertificate(name string) error {
	// Determine the current certificate status to decide which endpoint to use
	status, err := c.GetCertificate(name)

	if err != nil {
		return err
	}

	if status == nil {
		return fmt.Errorf("certificate %s not found", name)
	}

	var statusCode int

	switch status.State {
	case model.CertificateSigned:
		// If the certificate is signed, we can use the clean endpoint to revoke and clean in one step
		payload := struct {
			Certnames []string `json:"certnames"`
		}{
			Certnames: []string{name},
		}

		_, statusCode, err = c.call(http.MethodPut, "puppet-ca/v1/clean", payload, nil, nil)

	case model.CertificateRequested, model.CertificateRevoked:
		// If the certificate is revoked or requested, we must directly delete it
		_, statusCode, err = c.call(http.MethodDelete, fmt.Sprintf("puppet-ca/v1/certificate_status/%s", name), nil, nil, nil)

	default:
		return fmt.Errorf("certificate %s is in state %s, cannot clean", name, status.State)
	}

	if err != nil {
		return err
	}

	switch statusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	default:
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}
