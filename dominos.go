package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	LinkDominos      = `INSERT INTO "user" (basket, wii_id) VALUES ('[]', $1) ON CONFLICT(wii_id) DO UPDATE SET basket = '[]', wii_id = $1`
	UnlinkDominos    = `DELETE FROM "user" WHERE wii_id = $1`
	SetDominosLinked = `UPDATE users SET dominos_linked = $1 WHERE email = $2`
)

func handleDominos(c *gin.Context, query string, toggle bool) {
	/*
		email, ok := c.Get("email")
		if !ok {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": "some how got here unauthorized",
			})
			return
		}

		var wiiNumberStr string
		err := pool.QueryRow(ctx, GetWiiNumberUser, email.(string)).Scan(&wiiNumberStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		wiiNumber, err := strconv.ParseUint(wiiNumberStr, 10, 64)
		if err != nil {
			// Failed to parse Wii Number or invalid integer
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": "invalid Wii number",
			})
			return
		}

		number := nwc24.LoadWiiNumber(wiiNumber)
		if !number.CheckWiiNumber() {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": "invalid Wii number",
			})
			return
		}

		// Link the account now
		_, err = dominosPool.Exec(ctx, query, strconv.Itoa(int(number.GetHollywoodID())))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": err.Error(),
			})
			return
		}

		// Toggle linked flag
		_, err = pool.Exec(ctx, SetDominosLinked, toggle, email.(string))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": err.Error(),
			})
			return
		}*/
}

func linkDominos(c *gin.Context) {
	handleDominos(c, LinkDominos, true)
	c.Redirect(http.StatusFound, "/manage")
}

func unlinkDominos(c *gin.Context) {
	handleDominos(c, UnlinkDominos, false)
	c.Redirect(http.StatusFound, "/manage")
}
