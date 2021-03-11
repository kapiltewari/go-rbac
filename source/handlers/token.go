package handlers

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"go-rbac/source/dtos/request"
	"go-rbac/source/dtos/response"
	"go-rbac/source/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

//GenerateTokens ...
func (h *Handler) GenerateTokens(c *fiber.Ctx, role string, userID string) (string, string, error) {
	//ed25519 private key
	prb, _ := hex.DecodeString(os.Getenv("ACCESS_TOKEN_EDPRIVATE"))
	privateKey := ed25519.PrivateKey(prb)

	//access token claims
	accessTokenClaims := paseto.JSONToken{
		Subject:    userID,
		Audience:   role,
		Issuer:     os.Getenv("TITLE"),
		IssuedAt:   time.Now().Local(),
		Expiration: time.Now().Local().Add(time.Minute * 30), //30 minutes
	}

	//generate unique jti
	jti := uuid.New().String()

	//refresh token claims
	refreshTokenClaims := paseto.JSONToken{
		Subject:    userID,
		Jti:        jti,
		Audience:   role,
		Issuer:     os.Getenv("TITLE"),
		IssuedAt:   time.Now().Local(),
		Expiration: time.Now().Local().Add(time.Hour * 24 * 10), //10 days
	}

	//sign tokens with private key
	accessToken, err := paseto.NewV2().Sign(privateKey, accessTokenClaims, nil)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := paseto.NewV2().Sign(privateKey, refreshTokenClaims, nil)
	if err != nil {
		return "", "", err
	}

	//set jti as redis value refresh-token-user-id-%s
	rtRedis := fmt.Sprintf("refresh-token-user-id-%s", userID)
	if err := h.Redis.Set(c.Context(), rtRedis, jti, time.Duration(24*10*time.Hour)).Err(); err != nil {
		return "", "", err
	}

	//return
	return accessToken, refreshToken, nil
}

//ValidateAccessToken ...
func ValidateAccessToken(accessToken string) (paseto.JSONToken, error) {
	//ed25519 public key
	pub, _ := hex.DecodeString(os.Getenv("ACCESS_TOKEN_EDPUBLIC"))
	publicKey := ed25519.PublicKey(pub)

	//to extract claims
	var accessTokenClaims paseto.JSONToken

	//verify token with public key
	err := paseto.NewV2().Verify(accessToken, publicKey, &accessTokenClaims, nil)
	if err != nil {
		return paseto.JSONToken{}, err
	}

	//validate time
	validNow := paseto.ValidAt(time.Now().Local())
	err = validNow(&accessTokenClaims)
	if err != nil {
		return paseto.JSONToken{}, err
	}
	return accessTokenClaims, nil
}

//ValidateRefreshToken ...
func ValidateRefreshToken(refreshToken string) (paseto.JSONToken, error) {
	//ed25519 public key
	pub, _ := hex.DecodeString(os.Getenv("ACCESS_TOKEN_EDPUBLIC"))
	publicKey := ed25519.PublicKey(pub)

	//to extract claims
	var refreshTokenClaims paseto.JSONToken

	//verify token with public key
	err := paseto.NewV2().Verify(refreshToken, publicKey, &refreshTokenClaims, nil)
	if err != nil {
		return paseto.JSONToken{}, err
	}

	//validate time
	validNow := paseto.ValidAt(time.Now().Local())
	err = validNow(&refreshTokenClaims)
	if err != nil {
		return paseto.JSONToken{}, err
	}
	return refreshTokenClaims, nil
}

//RefreshTokens refreshes tokens
func (h *Handler) RefreshTokens(c *fiber.Ctx) error {
	var refreshToken string

	//check for browser cookie
	refreshToken = c.Cookies("refresh_token")

	//if empty string then check for refresh token in request body
	if refreshToken == "" {
		var req request.RefreshTokensRequest

		if err := c.BodyParser(&req); err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "")
		}

		errors := utils.ValidateStruct(req)
		if errors != nil {
			return utils.SendValidationError(c, errors)
		}

		//set refresh token
		refreshToken = req.RefreshToken
	}

	//validate refresh token
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "")
	}

	//get redis key refresh-token-user-id-%s
	jti, err := h.Redis.Get(c.Context(), fmt.Sprintf("refresh-token-user-id-%s", claims.Subject)).Result()
	if err != nil {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "")
	}

	//match jti
	if jti != claims.Jti {
		return utils.SendError(c, fiber.StatusUnauthorized, "")
	}

	//generate new tokens
	accessToken, refreshToken, err := h.GenerateTokens(c, claims.Audience, claims.Subject)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	return utils.SendResponse(c, fiber.StatusOK, &response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
