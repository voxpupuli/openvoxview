package middleware

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/sebastianrakel/openvoxview/config"
)

// SamlSP wraps the crewjam/saml ServiceProvider with thread-safe metadata refresh.
type SamlSP struct {
	mu sync.RWMutex
	sp saml.ServiceProvider
}

// SP returns a copy of the current ServiceProvider (safe for concurrent use).
func (s *SamlSP) SP() saml.ServiceProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sp
}

func (s *SamlSP) updateIDPMetadata(metadata *saml.EntityDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sp.IDPMetadata = metadata
}

// NewSamlServiceProvider creates a SAML SP from the given config.
// It loads the SP certificate, fetches/parses IdP metadata, and returns a ready-to-use SamlSP.
func NewSamlServiceProvider(cfg *config.SamlConfig) (*SamlSP, error) {
	// Load SP certificate and key
	keyPair, err := tls.LoadX509KeyPair(cfg.SpCertFile, cfg.SpKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load SAML SP certificate: %w", err)
	}

	leaf, err := x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse SAML SP certificate: %w", err)
	}

	// Parse IdP metadata
	idpMetadata, err := fetchIDPMetadata(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load IdP metadata: %w", err)
	}

	entityIDURL, err := url.Parse(cfg.SpEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SP entity ID URL: %w", err)
	}

	acsURL, err := url.Parse(cfg.SpAcsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SP ACS URL: %w", err)
	}

	signer, ok := keyPair.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("SAML SP private key does not implement crypto.Signer")
	}

	sp := saml.ServiceProvider{
		EntityID:          entityIDURL.String(),
		Key:               signer,
		Certificate:       leaf,
		AcsURL:            *acsURL,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: true,
	}

	ssp := &SamlSP{sp: sp}

	// Start background metadata refresh if using a URL
	if cfg.IdpMetadataURL != "" {
		go ssp.refreshMetadataLoop(cfg.IdpMetadataURL)
	}

	return ssp, nil
}

func fetchIDPMetadata(cfg *config.SamlConfig) (*saml.EntityDescriptor, error) {
	if cfg.IdpMetadataURL != "" {
		return fetchIDPMetadataFromURL(cfg.IdpMetadataURL)
	}
	if cfg.IdpMetadataFile != "" {
		return loadIDPMetadataFromFile(cfg.IdpMetadataFile)
	}
	return nil, fmt.Errorf("either idp_metadata_url or idp_metadata_file must be configured")
}

func fetchIDPMetadataFromURL(metadataURL string) (*saml.EntityDescriptor, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IdP metadata from %s: %w", metadataURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IdP metadata URL returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IdP metadata response: %w", err)
	}

	return parseIDPMetadata(data)
}

func loadIDPMetadataFromFile(path string) (*saml.EntityDescriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read IdP metadata file %s: %w", path, err)
	}
	return parseIDPMetadata(data)
}

func parseIDPMetadata(data []byte) (*saml.EntityDescriptor, error) {
	// Try parsing as EntityDescriptor first
	entity := &saml.EntityDescriptor{}
	if err := xml.Unmarshal(data, entity); err == nil && entity.IDPSSODescriptors != nil {
		return entity, nil
	}

	// Try parsing as EntitiesDescriptor (federation metadata)
	entities := &saml.EntitiesDescriptor{}
	if err := xml.Unmarshal(data, entities); err != nil {
		return nil, fmt.Errorf("failed to parse IdP metadata XML: %w", err)
	}

	for i := range entities.EntityDescriptors {
		if entities.EntityDescriptors[i].IDPSSODescriptors != nil {
			return &entities.EntityDescriptors[i], nil
		}
	}

	return nil, fmt.Errorf("no IdP entity found in metadata")
}

func (s *SamlSP) refreshMetadataLoop(metadataURL string) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		metadata, err := fetchIDPMetadataFromURL(metadataURL)
		if err != nil {
			log.Printf("WARNING: Failed to refresh IdP metadata: %v", err)
			continue
		}
		s.updateIDPMetadata(metadata)
		log.Printf("SAML: IdP metadata refreshed from %s", metadataURL)
	}
}

// GetAttribute extracts a named attribute value from a SAML assertion.
func GetAttribute(assertion *saml.Assertion, name string) string {
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if attr.Name == name || attr.FriendlyName == name {
				if len(attr.Values) > 0 {
					return attr.Values[0].Value
				}
			}
		}
	}
	return ""
}
