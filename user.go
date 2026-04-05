package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUserRequest(uid any) (map[string]any, error) {
	url := fmt.Sprintf("https://sso.riiconnect24.net/api/v3/core/users/%s/", uid)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.OIDCConfig.ServiceAccountToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response")
	}

	var result map[string]any
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user")
	}

	return result["attributes"].(map[string]any), nil
}

func getUser(c *gin.Context) {
	uid, _ := c.Get("uid")

	attrs, err := getUserRequest(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"attributes": attrs,
	})
}

func updatePublicProfile(c *gin.Context) {
	uid, _ := c.Get("uid")
	var req struct {
		PublicProfile bool `json:"public_profile"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	payload := map[string]any{
		"attributes": map[string]any{
			"public_profile": req.PublicProfile,
		},
	}

	err := updateUserRequest(uid, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
