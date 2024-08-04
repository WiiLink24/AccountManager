package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

const (
	checkShaQuery = `SELECT 1 FROM users WHERE`
)

func HomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

func NotLinkedPage(c *gin.Context) {
	c.HTML(http.StatusOK, "not_linked.html", nil)
}

func LinkHandler(c *gin.Context) {
	/* sha := c.Param("sha") */

	c.HTML(http.StatusOK, "linked.html", nil)
}