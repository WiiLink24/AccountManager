package main

import (
	"log"

	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	IsUserLinked = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
)

func HomePage(c *gin.Context) {
	if email, ok := c.Get("email"); ok {
		var exists bool
		err := pool.QueryRow(ctx, IsUserLinked, email.(string)).Scan(&exists)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		if exists {
			c.HTML(http.StatusOK, "home.html", nil)
		} else {
			c.HTML(http.StatusOK, "not_linked.html", gin.H{
				"email": email,
			})
		}

		return
	}

	c.HTML(http.StatusBadRequest, "error.html", gin.H{
		"Error": "some how got here unauthorized",
	})

}

func NotLinkedPage(c *gin.Context) {
	email, _ := c.Get("email")
	log.Printf("User with email %s is not linked", email)
	c.HTML(http.StatusOK, "not_linked.html", gin.H{
		"email": email,
	})
}
