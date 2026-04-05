package main

import (
	"log"
	"net/http"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
	username, _ := c.Get("username")
	email, _ := c.Get("email")
	wiis, _ := c.Get("wiis")

	exists := len(wiis.([]middleware.Wii)) != 0
	dominos := map[string]bool{}
	for _, wii := range wiis.([]middleware.Wii) {
		dominos[wii.WiiNumber] = wii.DominosLinked
	}

	if exists {
		log.Printf("User with username %s is linked!!!", username)

		if pfp, ok := c.Get("picture"); ok {
			c.HTML(http.StatusOK, "linked.html", gin.H{
				"username": username,
				"email":    email,
				"pfp":      pfp,
				"dominos":  dominos,
				"wiis":     wiis.([]middleware.Wii),
			})
		} else {
			c.HTML(http.StatusOK, "linked.html", gin.H{
				"username": username,
				"email":    email,
				"dominos":  dominos,
				"wiis":     wiis.([]middleware.Wii),
			})
		}
	} else {
		log.Printf("User with username %s is not linked", username)
		if pfp, ok := c.Get("picture"); ok {
			c.HTML(http.StatusOK, "not_linked.html", gin.H{
				"username": username,
				"pfp":      pfp,
				"email":    email,
			})
		} else {
			c.HTML(http.StatusOK, "not_linked.html", gin.H{
				"username": username,
				"email":    email,
			})
		}
	}
}

func NotLinkedPage(c *gin.Context) {
	email, _ := c.Get("email")
	log.Printf("User with email %s is not linked", email)
	c.HTML(http.StatusOK, "not_linked.html", gin.H{
		"email": email,
	})
}
