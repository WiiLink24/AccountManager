package main

import (
	"net/http"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/gin-gonic/gin"
)

func removeWii(c *gin.Context) {
	numberToRemove := c.PostForm("wiino")
	wiis, _ := c.Get("wiis")
	uid, _ := c.Get("uid")

	// Delete then update
	for i, wii := range wiis.([]middleware.Wii) {
		if wii.WiiNumber == numberToRemove {
			// Delete/DeleteFunc just zeros out the object which is not what we want to do.
			left := wiis.([]middleware.Wii)[0:i]
			right := wiis.([]middleware.Wii)[i+1:]
			wiis = append([]middleware.Wii{}, left...)
			wiis = append(wiis.([]middleware.Wii), right...)
		}
	}
	err := updateUserAttributes(uid, map[string]any{
		"wiis": wiis,
	})
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

	c.Redirect(http.StatusFound, "/manage")
}
