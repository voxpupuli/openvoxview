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

// PuppetTime wraps time.Time with JSON parsing that accepts both RFC 3339 / ISO 8601
// (used by the golang-puppet-ca implementation and expected by future implementations)
// and the named-timezone format (e.g. "2006-01-02T15:04:05UTC") used by the Clojure-based
// OpenVox Server, Puppet Server, OpenVoxDB, and PuppetDB implementations.
type PuppetTime struct {
	time.Time
}

// PuppetCaTime is an alias kept for backwards compatibility.
type PuppetCaTime = PuppetTime

// PuppetSerialNumber holds a certificate serial number that may arrive as either
// a JSON number (OpenVox Server / Puppet Server) or a JSON string (golang-puppet-ca
// and expected future implementations). It always marshals back as a JSON string.
type PuppetSerialNumber struct {
	val string
}

func (s *PuppetSerialNumber) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.val = str
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("cannot parse %s as a serial number: %w", data, err)
	}
	s.val = n.String()
	return nil
}

func (s PuppetSerialNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.val)
}

type CertificateStatus struct {
	Name         string            `json:"name"`
	State        CertificateState  `json:"state"`
	Fingerprint  string            `json:"fingerprint"`
	Fingerprints map[string]string `json:"fingerprints"`
	DnsAltNames  []string          `json:"dns_alt_names"`

	SubjectAltNames         *[]string              `json:"subject_alt_names,omitempty"`
	SerialNumber            *PuppetSerialNumber    `json:"serial_number,omitempty"`
	AuthorizationExtensions *map[string]string     `json:"authorization_extensions,omitempty"`
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

// puppetCaTimeLayout is the non-standard format used by OpenVox Server, Puppet Server,
// OpenVoxDB, and PuppetDB (Clojure implementations), which emit a named timezone
// abbreviation (e.g. "UTC", "EST") rather than the ISO 8601 "Z" suffix.
const puppetCaTimeLayout = "2006-01-02T15:04:05MST"

func (t *PuppetTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// golang-puppet-ca and future implementations use ISO 8601 / RFC 3339 ("Z" or "+HH:MM");
	// OpenVox Server / Puppet Server / OpenVoxDB / PuppetDB (Clojure) use a named timezone
	// abbreviation. Try the standard format first, then fall back.
	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		parsedTime, err = time.Parse(puppetCaTimeLayout, s)
		if err != nil {
			return fmt.Errorf("cannot parse %q as a Puppet CA timestamp: %w", s, err)
		}
	}

	t.Time = parsedTime
	return nil
}

func (t PuppetTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339))
}
