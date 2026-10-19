package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type CognitoClaims struct {
	TokenUse string `json:"token_use"`
	ClientID string `json:"client_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type Verifier struct {
	kf       keyfunc.Keyfunc
	issuer   string
	clientID string
}

type ctxKey struct{}

func NewVerifier(ctx context.Context, region string, poolID string, clientID string) (*Verifier, error) {
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, poolID)

	kf, err := keyfunc.NewDefaultCtx(ctx, []string{issuer + "/.well-known/jwks.json"})
	if err != nil {
		return nil, err
	}

	return &Verifier{kf: kf, issuer: issuer, clientID: clientID}, nil
}

func (v *Verifier) VerifyAccessToken(raw string) (*CognitoClaims, error) {
	claims := &CognitoClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, v.kf.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "access" {
		return nil, errors.New("not an access token")
	}
	if claims.ClientID != v.clientID {
		return nil, errors.New("wrong client id")
	}

	return claims, nil
}

func (v *Verifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		raw, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		claims, err := v.VerifyAccessToken(raw)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
