package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type Claims map[string]interface{}

func Parse(jwtSecret string, bearerToken string) (Claims, error) {
	if jwtSecret == "" {
		return nil, errors.New("No jwt secret")
	}

	// parse the token
	parsedToken, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		// validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected JWT signing method: %v", token.Header["alg"])
		}

		// return the signing key
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Infof("failed to parse JWT \"%s\": %v", bearerToken, err)
		return nil, err
	}

	// validate the token
	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		log.Infof("invalid auth token: %s", bearerToken)
		return nil, errors.New("invalid auth token")
	}

	claims := make(Claims)
	for k, v := range mapClaims {
		claims[k] = v
	}

	return claims, nil
}

func Create(tokenIssuer string, jwtSecret string, tokenLifetimeHours int, subjectID string, customClaims Claims) (string, error) {
	// ensure the user provided the secret signing key
	if jwtSecret == "" {
		return "", errors.New("No jwt secret")
	}

	// create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// prepare the JWT subject
	if subjectID == "" {
		subjectID = uuid.New().String()
	}

	// prepare the TTL of the token
	ttl := time.Hour * time.Duration(tokenLifetimeHours)

	// set standard token claims
	claims := token.Claims.(jwt.MapClaims)
	claims["iss"] = tokenIssuer                // issuer
	claims["aud"] = tokenIssuer                // audience
	claims["iat"] = time.Now().Unix()          // issued at
	claims["exp"] = time.Now().Add(ttl).Unix() // expiration time
	claims["jti"] = uuid.New().String()        // token ID
	claims["sub"] = subjectID                  // subject

	// set custom token claims
	for k, v := range customClaims {
		claims[k] = v
	}

	// sign the token using the secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	// display the token
	return tokenString, nil
}
