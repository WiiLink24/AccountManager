package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/WiiLink24/nwc24"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
	"strconv"
)

func linkDominos(c *gin.Context) {
	wiiNoStr := c.PostForm("dominos_wii_no")

	// Load and verify Wii Number
	wiiNoInt, err := strconv.ParseUint(wiiNoStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	wiiNo := nwc24.LoadWiiNumber(wiiNoInt)
	if !wiiNo.CheckWiiNumber() {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": fmt.Sprintf("WiiNumber %d is invalid", wiiNo),
		})
		return
	}

	// Now that we are verified, dial the socket and link.
	conn, err := net.Dial("unix", "/tmp/dominos.sock")
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	defer conn.Close()
	_, err = conn.Write([]byte(strconv.Itoa(int(wiiNo.GetHollywoodID()))))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Read response
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil && !errors.Is(err, io.EOF) {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	var result map[string]any
	err = json.Unmarshal(resp[:n-1], &result)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	if !result["success"].(bool) {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": result["error"].(string),
		})
	}

	// Finally we toggle linked in authentik API.
	wiis, _ := c.Get("wiis")
	wwfc, _ := c.Get("wwfc")
	dominos, _ := c.Get("dominos")
	uid, _ := c.Get("uid")
	tokenString, _ := c.Cookie("token")

	// Toggle the linkage
	if dominos.(map[string]bool)[wiiNoStr] {
		dominos.(map[string]bool)[wiiNoStr] = false
	} else {
		dominos.(map[string]bool)[wiiNoStr] = true
	}

	payload := map[string]any{
		"attributes": map[string]any{
			"wiis":    wiis,
			"wwfc":    wwfc,
			"dominos": dominos,
		},
	}

	err = updateUserRequest(uid, tokenString, payload)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
	}

	c.Redirect(http.StatusFound, "/manage")
}
