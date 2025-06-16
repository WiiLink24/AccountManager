package main

import (
	"bytes"
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
	url := fmt.Sprintf("https://sso.riiconnect24.net/api/v3/core/users/%s/", uid)
	tokenString, _ := c.Cookie("token")

	for i, m := range dominos.([]map[string]bool) {
		for k := range m {
			if k == wiiNoStr {
				dominos.([]map[string]bool)[i][k] = true
			}
		}
	}
	payload := map[string]any{
		"attributes": map[string]any{
			"wiis":    wiis,
			"wwfc":    wwfc,
			"dominos": dominos,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to marshal payload",
		})
	}

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to create request",
		})
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	httpResp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to send request",
		})
	}

	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	fmt.Println(string(body))

	c.Redirect(http.StatusFound, "/manage")
}
