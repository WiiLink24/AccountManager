package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/nwc24"
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
	wiiNumber := c.PostForm("wii_num")
	wwfcCert := c.PostForm("cert")

	// Verify cert first
	ngId, err := verifySignature("WIILINK_ACCOUNT_LINKER", wwfcCert)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Now that the certificate is verified, validate the Wii number.
	intWiiNumber, err := strconv.Atoi(wiiNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}

	wiiNoObj := nwc24.LoadWiiNumber(uint64(intWiiNumber))
	if !wiiNoObj.CheckWiiNumber() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid wii number",
		})
	}

	if wiiNoObj.GetHollywoodID() != ngId {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "device id associated with the Wii Number does not match the Wii.",
		})
	}

	// Add Wii number to list
	wiis, ok := c.Get("wiis")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get wii numbers",
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
			"error":   "failed to get WWFC accounts",
		})
	}

	if !slices.Contains(wwfc.([]string), strconv.Itoa(int(ngId))) {
		wwfc = append(wwfc.([]string), strconv.Itoa(int(ngId)))
	}

	// Finally for Dominos
	dominos, ok := c.Get("dominos")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get dominos data",
		})
	}

	if len(dominos.([]map[string]bool)) == 0 {
		dominos = append(dominos.([]map[string]bool), map[string]bool{wiiNumber: false})
	} else if !slices.ContainsFunc(dominos.([]map[string]bool), func(s map[string]bool) bool {
		// Dominos is a []map[string]bool
		for k, _ := range s {
			if k == wiiNumber {
				return true
			}
		}

		return false
	}) {
		dominos = append(dominos.([]map[string]bool), map[string]bool{wiiNumber: false})
	}

	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get uid",
		})
	}

	url := fmt.Sprintf("https://sso.riiconnect24.net/api/v3/core/users/%s/", uid)

	tokenString := c.GetHeader("Authorization")

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

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to send request",
		})
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"error":   "failed to update user",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
