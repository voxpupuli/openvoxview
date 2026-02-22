package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	CertificateRequested CertificateState = iota + 1
	CertificateSigned
	CertificateRevoked
)

type CertificateState uint8

var (
	CertificateStateName = map[CertificateState]string{
		CertificateRequested: "requested",
		CertificateSigned:    "signed",
		CertificateRevoked:   "revoked",
	}
	CertificateStateValue = map[string]CertificateState{
		"requested": CertificateRequested,
		"signed":    CertificateSigned,
		"revoked":   CertificateRevoked,
	}
)

func (s CertificateState) String() string {
	return CertificateStateName[s]
}

func ParseCertificateState(s string) (CertificateState, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	value, ok := CertificateStateValue[s]
	if !ok {
		return CertificateState(0), fmt.Errorf("%q is not a valid certificate state", s)
	}
	return CertificateState(value), nil
}

func (s CertificateState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *CertificateState) UnmarshalJSON(data []byte) (err error) {
	var state string
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	if *s, err = ParseCertificateState(state); err != nil {
		return err
	}
	return nil
}

func UniqueCertificateStates(states []CertificateState) []CertificateState {
	seen := make(map[CertificateState]struct{})
	result := make([]CertificateState, 0, len(states))

	for _, s := range states {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}

	return result
}

type PuppetCaTime struct {
	time.Time
}

type CertificateStatus struct {
	Name         string            `json:"name"`
	State        CertificateState  `json:"state"`
	Fingerprint  string            `json:"fingerprint"`
	Fingerprints map[string]string `json:"fingerprints"`
	DnsAltNames  []string          `json:"dns_alt_names"`

	SubjectAltNames         *[]string          `json:"subject_alt_names,omitempty"`
	SerialNumber            *int               `json:"serial_number,omitempty"`
	AuthorizationExtensions *map[string]string `json:"authorization_extensions,omitempty"`
	NotBefore               *PuppetCaTime      `json:"not_before,omitempty"`
	NotAfter                *PuppetCaTime      `json:"not_after,omitempty"`
}

type CertificateStatusQuery struct {
	States *[]CertificateState `json:"states"`
	Filter *string             `json:"filter"`
}

type CertificateStatusResponse struct {
	CertificateStatuses []CertificateStatus `json:"certificate_statuses"`
}

const puppetCaTimeLayout = "2006-01-02T15:04:05MST"

func (t *PuppetCaTime) UnmarshalJSON(data []byte) error {
	// Puppet CA returns time in a non-standard format, so we need to parse it manually ...
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsedTime, err := time.Parse(puppetCaTimeLayout, s)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

func (t PuppetCaTime) MarshalJSON() ([]byte, error) {
	// ... but when marshalling we want to use the standard RFC3339 format.
	return json.Marshal(t.Format(time.RFC3339))
}
