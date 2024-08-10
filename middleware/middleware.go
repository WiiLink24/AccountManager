package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			// We can't redirect off an Unauthorized status code.
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims, err := VerifyToken(tokenString)
		log.Printf("Claims: %v", claims)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if email, ok := claims["Email"].(string); ok {
			c.Set("email", email)
		} else {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if username, ok := claims["nickname"].(string); ok {
			c.Set("username", username)
			if pfp, ok := claims["picture"].(string); ok {
				c.Set("picture", pfp)
			} else {
				c.Set("picture", "")
			}
		} else {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
