package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/WiiLink24/nwc24"
	"github.com/gin-gonic/gin"
)

func updateUserRequest(uid any, payload map[string]any) error {
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.OIDCConfig.ServiceAccountToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update user")
	}

	return nil
}

func linkRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "https://sso.riiconnect24.net/device")
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
		return
	}

	wiiNoObj := nwc24.LoadWiiNumber(uint64(intWiiNumber))
	if !wiiNoObj.CheckWiiNumber() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid wii number",
		})
		return
	}

	if wiiNoObj.GetHollywoodID() != ngId {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "device id associated with the Wii Number does not match the Wii.",
		})
		return
	}

	_wiis, ok := c.Get("wiis")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get wiis",
		})
		return
	}

	wiis := _wiis.([]middleware.Wii)
	if slices.ContainsFunc(wiis, func(w middleware.Wii) bool {
		return w.WiiNumber == wiiNumber
	}) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "wii already linked",
		})
	}

	// Now add the object
	wii := middleware.Wii{
		WiiNumber:     wiiNumber,
		HollywoodID:   strconv.Itoa(int(ngId)),
		DominosLinked: false,
		JustEatLinked: false,
	}

	wiis = append(wiis, wii)

	uid, ok := c.Get("uid")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "failed to get uid",
		})
	}

	payload := map[string]any{
		"attributes": map[string]any{
			"wiis": wiis,
		},
	}

	err = updateUserRequest(uid, payload)
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
