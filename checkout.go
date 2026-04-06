package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func toggleProfile(c *gin.Context) {
	// Toggle public profile in authentik API.
	wiis, _ := c.Get("wiis")
	uid, _ := c.Get("uid")
	publicProfile, ok := c.Get("public_profile")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get public_profile",
		})
	}

	payload := map[string]any{
		"attributes": map[string]any{
			"public_profile": !publicProfile.(bool),
			"wiis":           wiis,
		},
	}

	err := updateUserRequest(uid, payload)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Refresh the token so our new changes are reflected on the page
	refreshToken, _ := c.Cookie("refresh_token")
	newToken, err := getNewToken(refreshToken)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.SetCookie("token", newToken.AccessToken, newToken.ExpiresIn, "", "", false, true)
	c.SetCookie("refresh_token", newToken.RefreshToken, newToken.ExpiresIn, "", "", false, true)

	c.Redirect(http.StatusFound, "/privacy")
}
