package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/appointly/appointly/backend/internal/domain/auth"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
)

// JWTService implements auth.TokenService using HMAC SHA-256 JWT tokens.
type JWTService struct {
	secretKey     []byte
	issuer        string
	tokenDuration time.Duration
}

// NewJWTService creates a new JWT token service.
func NewJWTService(secretKey string, issuer string, tokenDuration time.Duration) *JWTService {
	if tokenDuration == 0 {
		tokenDuration = 15 * time.Minute // default 15m access token
	}
	return &JWTService{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		tokenDuration: tokenDuration,
	}
}

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTClaims struct {
	Sub       string    `json:"sub"`
	Email     string    `json:"email"`
	OrgID     string    `json:"org_id,omitempty"`
	Role      string    `json:"role,omitempty"`
	IsAdmin   bool      `json:"is_admin,omitempty"`
	Iss       string    `json:"iss"`
	Iat       int64     `json:"iat"`
	Exp       int64     `json:"exp"`
}

func (s *JWTService) GenerateAccessToken(claims auth.Claims) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.tokenDuration)

	header := Header{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to marshal header: %w", err)
	}

	payload := JWTClaims{
		Sub:     claims.UserID.String(),
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
		Iss:     s.issuer,
		Iat:     now.Unix(),
		Exp:     expiresAt.Unix(),
	}
	if claims.OrgID != uuid.Nil {
		payload.OrgID = claims.OrgID.String()
	}
	if claims.Role != "" {
		payload.Role = claims.Role
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to marshal claims: %w", err)
	}

	b64Header := base64.RawURLEncoding.EncodeToString(headerJSON)
	b64Payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	unsignedToken := b64Header + "." + b64Payload
	signature := s.sign(unsignedToken)
	b64Signature := base64.RawURLEncoding.EncodeToString(signature)

	token := unsignedToken + "." + b64Signature
	return token, expiresAt, nil
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*auth.Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, apperror.TokenInvalid()
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSig := s.sign(unsignedToken)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, apperror.TokenInvalid()
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return nil, apperror.TokenInvalid()
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, apperror.TokenInvalid()
	}

	var jwtClaims JWTClaims
	if err := json.Unmarshal(payloadJSON, &jwtClaims); err != nil {
		return nil, apperror.TokenInvalid()
	}

	now := time.Now().Unix()
	if now >= jwtClaims.Exp {
		return nil, apperror.TokenExpired()
	}

	userID, err := uuid.Parse(jwtClaims.Sub)
	if err != nil {
		return nil, apperror.TokenInvalid()
	}

	var orgID uuid.UUID
	if jwtClaims.OrgID != "" {
		orgID, _ = uuid.Parse(jwtClaims.OrgID)
	}

	return &auth.Claims{
		UserID:    userID,
		Email:     jwtClaims.Email,
		OrgID:     orgID,
		Role:      jwtClaims.Role,
		IsAdmin:   jwtClaims.IsAdmin,
		ExpiresAt: time.Unix(jwtClaims.Exp, 0),
		IssuedAt:  time.Unix(jwtClaims.Iat, 0),
	}, nil
}

func (s *JWTService) sign(data string) []byte {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(data))
	return h.Sum(nil)
}
