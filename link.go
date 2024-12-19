package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	IsValidHash  = `SELECT EXISTS(SELECT 1 FROM accounts WHERE password = $1)`
	GetWiiNumber = `SELECT mlid FROM accounts WHERE password = $1`
	LinkAccount  = `INSERT INTO users (email, wii_number) VALUES ($1, $2)`
)

func link(c *gin.Context) {
	hash := c.Query("H")
	hash = strings.ToLower(hash)

	var valid bool
	err := wiiMailPool.QueryRow(ctx, IsValidHash, hash).Scan(&valid)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	if !valid {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"Error": "Invalid Wii console. Please ensure you are registered with WiiLink Wii Mail.",
		})
		return
	}

	// Link the account.
	email, ok := c.Get("email")
	if !ok {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "some how got here unauthorized",
		})
		return
	}

	var wiiNumber string
	err = wiiMailPool.QueryRow(ctx, GetWiiNumber, hash).Scan(&wiiNumber)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	_, err = pool.Exec(ctx, LinkAccount, email.(string), wiiNumber)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "linked.html", gin.H{
		"message": "Successfully linked your Wii to your account!",
	})
}
