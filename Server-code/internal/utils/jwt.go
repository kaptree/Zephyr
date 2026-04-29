package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"time"

	"labelpro-server/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	jwtCfg     *config.JWTConfig
)

func InitJWT(cfg *config.JWTConfig) error {
	jwtCfg = cfg

	privData, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return err
	}

	privBlock, _ := pem.Decode(privData)
	if privBlock == nil {
		return errors.New("failed to parse private key PEM")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
		if err2 != nil {
			return errors.New("failed to parse private key")
		}
		var ok bool
		privKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return errors.New("private key is not RSA")
		}
	}
	privateKey = privKey

	pubData, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return err
	}

	pubBlock, _ := pem.Decode(pubData)
	if pubBlock == nil {
		return errors.New("failed to parse public key PEM")
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		pubCert, err2 := x509.ParsePKCS1PublicKey(pubBlock.Bytes)
		if err2 != nil {
			return errors.New("failed to parse public key")
		}
		pubKey = pubCert
	}
	var ok bool
	publicKey, ok = pubKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("public key is not RSA")
	}

	return nil
}

func GenerateTokenPair(userID, username, role, departmentID string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		DepartmentID: departmentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtCfg.AccessTokenExpireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jwtCfg.Issuer,
			ID:        uuid.New().String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(privateKey)
	if err != nil {
		return nil, err
	}

	refreshClaims := &Claims{
		UserID:       userID,
		Username:     username,
		Role:         role,
		DepartmentID: departmentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtCfg.RefreshTokenExpireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jwtCfg.Issuer,
			ID:        uuid.New().String(),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(privateKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(jwtCfg.AccessTokenExpireSeconds),
	}, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GetAccessTokenExpiry() time.Duration {
	return time.Duration(jwtCfg.AccessTokenExpireSeconds) * time.Second
}
