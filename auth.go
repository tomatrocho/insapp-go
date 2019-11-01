package insapp

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// TokenClaims is the JWT encoding format.
type TokenClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

const (
	refreshTokenValidTime = time.Hour * 72
	authTokenValidTime    = time.Minute * 15
)

// InitJWT reads the key files before starting http handlers
func InitJWT(config Config) error {
	signBytes, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		return err
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}

	verifyBytes, err := ioutil.ReadFile(config.PublicKeyPath)
	if err != nil {
		return err
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return err
	}

	return nil
}

// CreateNewTokens creates auth and refresh tokens
func CreateNewTokens(username string, role string) (string, string, error) {
	// generate the auth token
	authTokenString, err := createAuthTokenString(username, role)
	if err != nil {
		return "", "", err
	}

	// generate the refresh token
	refreshTokenString, err := createRefreshTokenString(username, role)
	if err != nil {
		return "", "", err
	}

	return authTokenString, refreshTokenString, nil
}

// createAuthTokenString creates an auth token
func createAuthTokenString(username string, role string) (string, error) {
	authTokenExpiration := time.Now().Add(authTokenValidTime).Unix()

	authClaims := TokenClaims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: authTokenExpiration,
		},
	}

	// Create a signer for rsa 256
	authJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), authClaims)

	// Generate the auth token string
	return authJwt.SignedString(signKey)
}

// createRefreshTokenString create a refresh token
func createRefreshTokenString(username string, role string) (string, error) {
	refreshTokenExpiration := time.Now().Add(refreshTokenValidTime).Unix()

	token := StoreRefreshToken()

	refreshClaims := TokenClaims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			Id:        token.JTI,
			ExpiresAt: refreshTokenExpiration,
		},
	}

	// Create a signer for rsa 256
	refreshJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), refreshClaims)

	// Generate the refresh token string
	return refreshJwt.SignedString(signKey)
}

// CheckAndRefreshTokens renews the auth token if needed
func CheckAndRefreshTokens(authTokenString string, refreshTokenString string) (string, string, error) {
	var newAuthTokenString string
	var newRefreshTokenString string

	// Check that it matches what's in the auth token claims
	authToken, err := jwt.ParseWithClaims(authTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	// The auth token is still valid
	if _, ok := authToken.Claims.(*TokenClaims); ok && authToken.Valid {
		// Update the expiration time of refresh token, but don't save to the database
		// We don't need to check if our refresh token is valid here
		// because we aren't renewing the auth token
		newRefreshTokenString, err = updateRefreshTokenExpiration(refreshTokenString)

		return newRefreshTokenString, authTokenString, nil
	}

	if ve, ok := err.(*jwt.ValidationError); ok {
		// The auth token has expired
		if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			newAuthTokenString, err = updateAuthTokenString(authTokenString, refreshTokenString)
			if err != nil {
				return "", "", err
			}

			// Update the expiration time of refresh token string
			newRefreshTokenString, err = updateRefreshTokenExpiration(refreshTokenString)
			if err != nil {
				return "", "", err
			}

			return newRefreshTokenString, newAuthTokenString, nil
		}
	}

	return "", "", err
}

func updateAuthTokenString(authTokenString string, refreshTokenString string) (string, error) {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	refreshTokenClaims, ok := refreshToken.Claims.(*TokenClaims)
	if !ok {
		return "", err
	}

	// Check if the refresh token has been revoked
	if CheckRefreshToken(refreshTokenClaims.StandardClaims.Id) {
		// The refresh token has not been revoked
		// Has it expired?
		if refreshToken.Valid {
			// The refresh token has not expired
			// Issue a new auth token
			authToken, err := jwt.ParseWithClaims(authTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
				return verifyKey, nil
			})

			authTokenClaims, ok := authToken.Claims.(*TokenClaims)
			if !ok {
				return "", err
			}

			return createAuthTokenString(authTokenClaims.Username, authTokenClaims.Role)
		}

		// The refresh token has expired!
		// Revoke the token in our db and require the user to login again
		DeleteRefreshToken(refreshTokenClaims.StandardClaims.Id)

		return "", errors.New("Unauthorized")
	}

	// The refresh token has been revoked!
	return "", errors.New("Unauthorized")
}

func updateRefreshTokenExpiration(refreshTokenString string) (string, error) {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	refreshTokenClaims, ok := refreshToken.Claims.(*TokenClaims)
	if !ok {
		return "", err
	}

	refreshTokenExpiration := time.Now().Add(refreshTokenValidTime).Unix()

	refreshClaims := TokenClaims{
		Role: refreshTokenClaims.Role,
		StandardClaims: jwt.StandardClaims{
			Id:        refreshTokenClaims.StandardClaims.Id,
			ExpiresAt: refreshTokenExpiration,
		},
	}

	// Create a signer for rsa 256
	refreshJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), refreshClaims)

	// Generate the refresh token string
	return refreshJwt.SignedString(signKey)
}

func RevokeRefreshToken(refreshTokenString string) error {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		return errors.New("Could not parse refresh token with claims")
	}

	refreshTokenClaims, ok := refreshToken.Claims.(*TokenClaims)
	if !ok {
		return errors.New("Could not read refresh token claims")
	}

	DeleteRefreshToken(refreshTokenClaims.StandardClaims.Id)

	return nil
}
