package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	IsUserLinked     = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	GetWiiNumberUser = `SELECT wii_number FROM users WHERE email = $1`
	GetLinkedDominos = `SELECT dominos_linked FROM users WHERE email = $1`
	IsProfilePublic  = `SELECT public_profile FROM users WHERE email = $1`
)

func HomePage(c *gin.Context) {
	username, _ := c.Get("username")
	email, _ := c.Get("email")
	wiis, _ := c.Get("wiis")
	dominos, _ := c.Get("dominos")
	publicProfile, _ := c.Get("public_profile")

	exists := len(wiis.([]string)) != 0

	if exists {
		log.Printf("User with username %s is linked!!!", username)

		if pfp, ok := c.Get("picture"); ok {
			c.HTML(http.StatusOK, "linked.html", gin.H{
				"username":       username,
				"email":          email,
				"pfp":            pfp,
				"wiiNumbers":     wiis.([]string),
				"dominos":        dominos.(map[string]bool),
				"public_profile": publicProfile,
			})
		} else {
			c.HTML(http.StatusOK, "linked.html", gin.H{
				"username":       username,
				"email":          email,
				"wiiNumbers":     wiis.([]string),
				"dominos":        dominos.(map[string]bool),
				"public_profile": publicProfile,
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
