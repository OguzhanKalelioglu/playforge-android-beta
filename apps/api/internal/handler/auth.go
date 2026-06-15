package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/service"
)

type AuthHandler struct {
	auth   *service.AuthService
	logger *zap.Logger
}

func NewAuthHandler(auth *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{auth: auth, logger: logger}
}

func (h *AuthHandler) Register(r *gin.Engine) {
	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", h.Register_)
		api.POST("/login", h.Login)
		api.POST("/refresh", h.Refresh)
		api.POST("/logout", h.Logout)
		api.GET("/me", middleware.AuthRequiredJWT(), h.Me)
	}
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register_(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, password required"})
		return
	}

	u, err := h.auth.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		default:
			h.logger.Error("register failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
		"role":  u.Role,
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		h.logger.Error("login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.auth.WriteAuthCookies(c.Writer, result)
	c.JSON(http.StatusOK, gin.H{
		"id":       result.User.ID,
		"email":    result.User.Email,
		"name":     result.User.Name,
		"role":     result.User.Role,
		"access_token": result.AccessToken,
		"expires_at":   result.ExpiresAt,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	// Hem body'den hem cookie'den kabul et
	refresh := ""
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		refresh = cookie
	}
	if refresh == "" {
		var req refreshReq
		_ = c.ShouldBindJSON(&req)
		refresh = req.RefreshToken
	}
	if refresh == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), refresh)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	h.auth.WriteAuthCookies(c.Writer, result)
	c.JSON(http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"expires_at":   result.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		_ = h.auth.Logout(c.Request.Context(), cookie)
	}
	h.auth.WriteClearCookies(c.Writer)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uidStr, _ := c.Get(middleware.CtxUserID)
	uid, err := uuid.Parse(uidStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}
	u, err := h.auth.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
		"role":  u.Role,
	})
}
