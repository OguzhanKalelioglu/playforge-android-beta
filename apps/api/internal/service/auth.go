package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/testerscommunity/api/internal/lib"
	"github.com/testerscommunity/api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrWeakPassword       = errors.New("password too weak")
	ErrSessionRevoked     = errors.New("session revoked")
	ErrSessionExpired     = errors.New("session expired")
)

type AuthService struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
	jwt      *lib.JWTManager
}

func NewAuthService(u *repository.UserRepository, s *repository.SessionRepository, j *lib.JWTManager) *AuthService {
	return &AuthService{users: u, sessions: s, jwt: j}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*repository.User, error) {
	if len(in.Password) < 8 {
		return nil, ErrWeakPassword
	}

	// Email zaten kayıtlı mı?
	if existing, _ := s.users.GetByEmail(ctx, in.Email); existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &repository.User{
		Email:         in.Email,
		PasswordHash:  string(hash),
		Name:          in.Name,
		Role:          "customer",
		EmailVerified: false,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         *repository.User
}

func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ip string) (*LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	access, expires, err := s.jwt.GenerateAccess(u.ID.String(), u.Email, u.Role)
	if err != nil {
		return nil, err
	}

	refresh, hash, err := lib.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}

	ua := userAgent
	ipAddr := ip
	sess := &repository.Session{
		UserID:           u.ID,
		RefreshTokenHash: hash,
		UserAgent:        &ua,
		IPAddress:        &ipAddr,
		ExpiresAt:        time.Now().Add(lib.RefreshTokenTTL),
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expires,
		User:         u,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	hash := lib.HashToken(refreshToken)
	sess, err := s.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if sess.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	access, expires, err := s.jwt.GenerateAccess(u.ID.String(), u.Email, u.Role)
	if err != nil {
		return nil, err
	}
	_ = s.sessions.Touch(ctx, sess.ID)

	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refreshToken, // aynısı
		ExpiresAt:    expires,
		User:         u,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := lib.HashToken(refreshToken)
	sess, err := s.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil
	}
	return s.sessions.Revoke(ctx, sess.ID)
}

func (s *AuthService) SetCookies(c interface{ SetCookie(...any) }, result *LoginResult) {
	// type-safe version aşağıda
}

func (s *AuthService) WriteAuthCookies(w http.ResponseWriter, result *LoginResult) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.ExpiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(lib.RefreshTokenTTL),
	})
}

func (s *AuthService) WriteClearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/", MaxAge: -1})
}

func (s *AuthService) GetByID(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	return s.users.GetByID(ctx, id)
}
