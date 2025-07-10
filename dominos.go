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
	"net/url"
	"strconv"
	"strings"
)

type refreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func getNewToken(refreshToken string) (*refreshResult, error) {
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", refreshToken)
	formData.Set("client_id", config.OIDCConfig.ClientID)
	formData.Set("client_secret", config.OIDCConfig.ClientSecret)

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://sso.riiconnect24.net/application/o/token/", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result refreshResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

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
		return
	}

	// Finally we toggle linked in authentik API.
	wiis, _ := c.Get("wiis")
	wwfc, _ := c.Get("wwfc")
	dominos, _ := c.Get("dominos")
	uid, _ := c.Get("uid")

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

	err = updateUserRequest(uid, payload)
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
