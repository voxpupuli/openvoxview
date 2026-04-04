package handler

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"github.com/sebastianrakel/openvoxview/config"
	"github.com/sebastianrakel/openvoxview/db"
	"github.com/sebastianrakel/openvoxview/middleware"
)

type AuthHandler struct {
	config      *config.Config
	database    *db.Database
	rateLimiter *rateLimiter
	samlSP      *middleware.SamlSP
}

func NewAuthHandler(config *config.Config, database *db.Database) *AuthHandler {
	return &AuthHandler{
		config:      config,
		database:    database,
		rateLimiter: newRateLimiter(),
	}
}

func NewAuthHandlerWithSAML(config *config.Config, database *db.Database, samlSP *middleware.SamlSP) *AuthHandler {
	return &AuthHandler{
		config:      config,
		database:    database,
		rateLimiter: newRateLimiter(),
		samlSP:      samlSP,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type createUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

type updateUserRequest struct {
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
	IsAdmin     *bool   `json:"is_admin"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	clientIP := c.ClientIP()
	if !h.rateLimiter.allow(clientIP) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, NewErrorResponse(errors.New("too many login attempts, try again later")))
		return
	}

	user, err := h.database.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, db.ErrInvalidCredentials) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewErrorResponse(errors.New("invalid username or password")))
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	tokens, err := h.issueTokenPair(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	log.Printf("[AUDIT] User logged in: %s", user.Username)
	c.JSON(http.StatusOK, NewSuccessResponse(tokens))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	rt, err := h.database.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, NewErrorResponse(errors.New("invalid or expired refresh token")))
		return
	}

	// Revoke old token (rotation)
	h.database.RevokeRefreshToken(req.RefreshToken)

	user, err := h.database.GetUserByID(rt.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, NewErrorResponse(errors.New("user not found")))
		return
	}

	tokens, err := h.issueTokenPair(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(tokens))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	h.database.RevokeRefreshToken(req.RefreshToken)

	username, _ := c.Get("username")
	log.Printf("[AUDIT] User logged out: %v", username)

	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	user, err := h.database.GetUserByID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(user))
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.database.ListUsers()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, NewSuccessResponse(users))
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	user, err := h.database.CreateUser(req.Username, req.Email, req.DisplayName, req.Password, req.IsAdmin)
	if err != nil {
		if errors.Is(err, db.ErrUsernameExists) {
			c.AbortWithStatusJSON(http.StatusConflict, NewErrorResponse(err))
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	log.Printf("[AUDIT] User created: %s (by %v)", req.Username, c.GetString("username"))
	c.JSON(http.StatusCreated, NewSuccessResponse(user))
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(errors.New("invalid user id")))
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	// Self-demote guard: prevent admin from removing their own admin flag
	currentUserIDStr, _ := c.Get("user_id")
	currentUserID, _ := strconv.ParseInt(currentUserIDStr.(string), 10, 64)
	if id == currentUserID && req.IsAdmin != nil && !*req.IsAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, NewErrorResponse(errors.New("cannot remove your own admin role")))
		return
	}

	user, err := h.database.UpdateUser(id, req.Email, req.DisplayName, req.Password, req.IsAdmin)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(err))
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	log.Printf("[AUDIT] User updated: id=%d (by %v)", id, c.GetString("username"))
	c.JSON(http.StatusOK, NewSuccessResponse(user))
}

func (h *AuthHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(errors.New("invalid user id")))
		return
	}

	// Prevent self-deletion
	currentUserIDStr, _ := c.Get("user_id")
	currentUserID, _ := strconv.ParseInt(currentUserIDStr.(string), 10, 64)
	if id == currentUserID {
		c.AbortWithStatusJSON(http.StatusForbidden, NewErrorResponse(errors.New("cannot delete your own account")))
		return
	}

	if err := h.database.DeleteUser(id); err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(err))
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	log.Printf("[AUDIT] User deleted: id=%d (by %v)", id, c.GetString("username"))
	c.JSON(http.StatusOK, NewSuccessResponse(nil))
}

func (h *AuthHandler) issueTokenPair(user *db.User) (*loginResponse, error) {
	accessToken, expiresAt, err := middleware.GenerateAccessToken(
		user.ID, user.Username, user.Email, user.DisplayName, user.IsAdmin,
		h.config.Auth.JwtSecret, h.config.Auth.AccessTokenTTL,
	)
	if err != nil {
		return nil, err
	}

	rawRefresh, hashRefresh, err := db.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshExpiry := time.Now().Add(time.Duration(h.config.Auth.RefreshTokenTTL) * 24 * time.Hour)
	if err := h.database.StoreRefreshToken(user.ID, hashRefresh, refreshExpiry); err != nil {
		return nil, err
	}

	expiresIn := expiresAt - time.Now().Unix()

	return &loginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    expiresIn,
	}, nil
}

// SamlMetadata returns the SP metadata XML for IdP registration.
func (h *AuthHandler) SamlMetadata(c *gin.Context) {
	if h.samlSP == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(errors.New("SAML not configured")))
		return
	}
	sp := h.samlSP.SP()
	metadata := sp.Metadata()
	data, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.Data(http.StatusOK, "application/xml", data)
}

// SamlLogin initiates SAML SSO by redirecting to the IdP.
func (h *AuthHandler) SamlLogin(c *gin.Context) {
	if h.samlSP == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(errors.New("SAML not configured")))
		return
	}
	sp := h.samlSP.SP()

	authnRequest, err := sp.MakeAuthenticationRequest(
		sp.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("failed to create SAML AuthnRequest: %w", err)))
		return
	}

	// Store the request ID in a cookie so ACS can validate InResponseTo
	c.SetCookie("saml_request_id", authnRequest.ID, 300, "/", "", false, true)

	redirectURL, err := authnRequest.Redirect("", &sp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(fmt.Errorf("failed to build SAML redirect URL: %w", err)))
		return
	}

	c.Redirect(http.StatusFound, redirectURL.String())
}

// SamlACS handles the SAML Assertion Consumer Service callback.
func (h *AuthHandler) SamlACS(c *gin.Context) {
	if h.samlSP == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(errors.New("SAML not configured")))
		return
	}

	sp := h.samlSP.SP()

	err := c.Request.ParseForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(fmt.Errorf("failed to parse form: %w", err)))
		return
	}

	// Retrieve the request ID from the cookie set during SamlLogin
	var possibleRequestIDs []string
	if requestID, err := c.Cookie("saml_request_id"); err == nil && requestID != "" {
		possibleRequestIDs = append(possibleRequestIDs, requestID)
	}

	// Clear the cookie
	c.SetCookie("saml_request_id", "", -1, "/", "", false, true)

	assertion, err := sp.ParseResponse(c.Request, possibleRequestIDs)
	if err != nil {
		// crewjam/saml hides the real error in InvalidResponseError.PrivateErr
		var ire *saml.InvalidResponseError
		if errors.As(err, &ire) {
			log.Printf("[SAML] ACS validation failed: %v (detail: %v)", err, ire.PrivateErr)
		} else {
			log.Printf("[SAML] ACS validation failed: %v", err)
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, NewErrorResponse(fmt.Errorf("SAML assertion validation failed: %w", err)))
		return
	}

	// Extract attributes
	email := middleware.GetAttribute(assertion, h.config.Auth.Saml.AttrEmail)
	givenName := middleware.GetAttribute(assertion, h.config.Auth.Saml.AttrGivenName)
	surname := middleware.GetAttribute(assertion, h.config.Auth.Saml.AttrSurname)
	displayName := middleware.GetAttribute(assertion, h.config.Auth.Saml.AttrDisplayName)

	if email == "" {
		log.Printf("[SAML] ACS: no email attribute in assertion")
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(errors.New("SAML assertion missing required email attribute")))
		return
	}

	// Upsert user in database
	user, err := h.database.UpsertSamlUser(email, givenName, surname, displayName)
	if err != nil {
		log.Printf("[SAML] ACS: failed to upsert user %s: %v", email, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	// Issue token pair (same as local login)
	tokens, err := h.issueTokenPair(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	log.Printf("[AUDIT] SAML user logged in: %s", user.Username)

	// Redirect to frontend with tokens in query params
	redirectURL := fmt.Sprintf("/ui/?#/login?token=%s&refresh=%s", tokens.AccessToken, tokens.RefreshToken)
	c.Redirect(http.StatusFound, redirectURL)
}

// Simple in-memory rate limiter: 5 attempts per IP per minute
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	window := now.Add(-1 * time.Minute)

	// Clean old entries
	valid := make([]time.Time, 0)
	for _, t := range rl.attempts[ip] {
		if t.After(window) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= 5 {
		rl.attempts[ip] = valid
		return false
	}

	rl.attempts[ip] = append(valid, now)
	return true
}
