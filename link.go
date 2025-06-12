package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
	"strconv"
)

const (
	IsValidHash  = `SELECT EXISTS(SELECT 1 FROM accounts WHERE password = $1)`
	GetWiiNumber = `SELECT mlid FROM accounts WHERE password = $1`
)

func link(c *gin.Context) {
	hash := c.PostForm("wii_num")
	wwfc_cert := c.PostForm("cert")

	// Verify cert first
	ngId, err := verifySignature("WIILINK_ACCOUNT_LINKER", wwfc_cert)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	fmt.Println(ngId)
	// Now verify Wii Number
	var valid bool
	err = wiiMailPool.QueryRow(ctx, IsValidHash, hash).Scan(&valid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid Wii console. Please ensure your Wii is registered with Wii Mail.",
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

	// Add Wii number to list
	wiis, ok := c.Get("wiis")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get wii numbers",
		})
	}

	if !slices.Contains(wiis.([]string), wiiNumber) {
		wiis = append(wiis.([]string), wiiNumber)
	}

	// Same for wwfc
	wwfc, ok := c.Get("wwfc")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get WWFC accounts",
		})
	}

	if !slices.Contains(wwfc.([]string), strconv.Itoa(int(ngId))) {
		wwfc = append(wwfc.([]string), strconv.Itoa(int(ngId)))
	}

	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to get uid",
		})
	}

	url := fmt.Sprintf("https://sso.riiconnect24.net/api/v3/core/users/%s/", uid)

	tokenString := c.GetHeader("Authorization")

	payload := map[string]any{
		"attributes": map[string]any{
			"wiis": wiis,
			"wwfc": wwfc,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to marshal payload",
		})
	}

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to create request",
		})
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "failed to send request",
		})
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "failed to update user",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
