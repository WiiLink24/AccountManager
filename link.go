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

func updateUserRequest(uid any, authorization string, payload map[string]any) error {
	url := fmt.Sprintf("https://sso.riiconnect24.net/api/v3/core/users/%s/", uid)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload")
	}

	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorization))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update user")
	}

	return nil
}

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

	dMap := dominos.(map[string]bool)
	if len(dMap) == 0 {
		dMap = map[string]bool{wiiNumber: false}
	} else if _, ok := dMap[wiiNumber]; !ok {
		// Dictionary is not empty, and the current Wii Number is not present.
		dMap[wiiNumber] = false
	}

	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get uid",
		})
	}

	tokenString := c.GetHeader("Authorization")

	payload := map[string]any{
		"attributes": map[string]any{
			"wiis":    wiis,
			"wwfc":    wwfc,
			"dominos": dMap,
		},
	}

	err = updateUserRequest(uid, tokenString, payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
