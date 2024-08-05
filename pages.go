package main

import (
	"log"
	
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	IsUserLinked = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
)

func HomePage(c *gin.Context) {
	if email, ok := c.Get("email"); ok {
        log.Printf("Checking if user with email %s is linked", email)
        var exists bool
        err := pool.QueryRow(ctx, IsUserLinked, email.(string)).Scan(&exists)
        
        if err != nil {
            log.Printf("Error querying database: %v", err)
            c.HTML(http.StatusBadRequest, "error.html", gin.H{
                "Error": err.Error(),
            })
            return
        }

        if exists {
            log.Printf("User with email %s is linked", email)
            c.HTML(http.StatusOK, "home.html", nil)
        } else {
            log.Printf("User with email %s is not linked", email)
            c.HTML(http.StatusOK, "not_linked.html", nil)
        }

        return
	}

	c.HTML(http.StatusBadRequest, "error.html", gin.H{
		"Error": "some how got here unauthorized",
	})

}

func NotLinkedPage(c *gin.Context) {
	c.HTML(http.StatusOK, "not_linked.html", nil)
}

func LinkHandler(c *gin.Context) {
	/* sha := c.Param("sha") */
	c.HTML(http.StatusOK, "linked.html", nil)
}
