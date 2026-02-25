package handler

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/model"
	"github.com/sebastianrakel/openvoxview/puppetca"
	"github.com/sebastianrakel/openvoxview/puppetdb"
)

type CaHandler struct {
	config   *config.Config
	caClient *puppetca.Client
}

func NewCaHandler(config *config.Config) *CaHandler {
	return &CaHandler{
		config:   config,
		caClient: puppetca.NewClient(config),
	}
}

func (h *CaHandler) QueryCertificateStatuses(c *gin.Context) {
	var query model.CertificateStatusQuery

	if err := c.ShouldBindJSON(&query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	if query.States != nil {
		states := model.UniqueCertificateStates(*query.States)

		if len(states) == len(model.CertificateStateValue) {
			// all states are requested, so we can ignore the filter and just get all certs
			query.States = nil
		} else {
			query.States = &states
		}
	}

	resultCerts := make([]model.CertificateStatus, 0)

	if query.States != nil {
		for _, state := range *query.States {
			certs, err := h.caClient.GetCertificates(&state)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
				return
			}
			resultCerts = append(resultCerts, certs...)
		}
	} else {
		certs, err := h.caClient.GetCertificates(nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
			return
		}
		resultCerts = certs
	}

	if query.Filter != nil {
		resultCerts = slices.Collect(func(yield func(model.CertificateStatus) bool) {
			for _, cert := range resultCerts {
				if strings.Contains(cert.Name, *query.Filter) ||
					strings.Contains(cert.Fingerprint, *query.Filter) ||
					slices.Contains(cert.DnsAltNames, *query.Filter) {
					if !yield(cert) {
						return
					}
				}
			}
		})
	}

	response := model.CertificateStatusResponse{
		CertificateStatuses: resultCerts,
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

func (h *CaHandler) SignCertificate(c *gin.Context) {
	name := c.Param("name")

	log.Printf("[AUDIT] CA Signing: %s", name)

	err := h.caClient.SignCertificate(name)

	if err != nil {
		log.Printf("Error signing certificate: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (h *CaHandler) RevokeCertificate(c *gin.Context) {
	name := c.Param("name")

	log.Printf("[AUDIT] CA Revoking: %s", name)

	err := h.caClient.RevokeCertificate(name)

	if err != nil {
		log.Printf("Error revoking certificate: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	err = h.deactivateNode(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (h *CaHandler) CleanCertificate(c *gin.Context) {
	name := c.Param("name")

	log.Printf("[AUDIT] CA Cleaning: %s", name)

	err := h.caClient.CleanCertificate(name)

	if err != nil {
		log.Printf("Error cleaning certificate: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	err = h.deactivateNode(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (h *CaHandler) deactivateNode(certname string) error {
	if !h.config.PuppetCA.DeactivateNodes {
		return nil
	}

	log.Printf("[AUDIT] Deactivating node: %s", certname)

	pdb := puppetdb.NewClient()
	resp, err := pdb.DeactivateNode(certname)

	if err != nil {
		log.Printf("Error deactivating node: %s", err)
	} else {
		log.Printf("Deactivated node %s with command UUID: %s", certname, resp.Uuid)
	}

	return err
}
