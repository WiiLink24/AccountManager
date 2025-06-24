package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthenticationMiddleware(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Verify the OpenID Connect idToken.
		ctx := context.Background()
		idToken, err := verifier.Verify(ctx, tokenString)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Parse custom claims if needed.
		var claims struct {
			UserId   string          `json:"sub"`
			Email    string          `json:"email"`
			Username string          `json:"preferred_username"`
			Wiis     []string        `json:"wiis"`
			WWFC     []string        `json:"wwfc"`
			Dominos  map[string]bool `json:"dominos"`
		}

		if err = idToken.Claims(&claims); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		if claims.Email != "" {
			c.Set("email", claims.Email)
			c.Set("picture", fmt.Sprintf("%x", sha256.Sum256([]byte(claims.Email))))
		} else {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		if claims.Username != "" {
			c.Set("username", claims.Username)
		} else {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("uid", claims.UserId)
		c.Set("wiis", claims.Wiis)
		c.Set("wwfc", claims.WWFC)
		c.Set("dominos", claims.Dominos)
		c.Next()
	}
}

func AuthenticationPOSTMiddleware(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// We can't redirect off an Unauthorized status code.
			c.Status(http.StatusBadRequest)
			c.Abort()
			return
		}

		// Verify the OpenID Connect idToken.
		ctx := context.Background()
		idToken, err := verifier.Verify(ctx, tokenString)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Parse custom claims if needed.
		var claims struct {
			Email    string          `json:"email"`
			Username string          `json:"preferred_username"`
			Name     string          `json:"name"`
			UserId   string          `json:"sub"`
			Groups   []string        `json:"groups"`
			Wiis     []string        `json:"wiis"`
			WWFC     []string        `json:"wwfc"`
			Dominos  map[string]bool `json:"dominos"`
		}

		if err = idToken.Claims(&claims); err != nil {
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Set("uid", claims.UserId)
		c.Set("wiis", claims.Wiis)
		c.Set("wwfc", claims.WWFC)
		c.Set("dominos", claims.Dominos)
		c.Next()
	}
}
