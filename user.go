package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/logrusorgru/aurora/v4"

	"io"
	"log"
	"net/http"
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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(aurora.Red("error closing body:"), err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response")
	}

	var result map[string]any
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user")
	}

	attributes, ok := result["attributes"].(map[string]any)
	if !ok {
		// User has no attributes set yet.
		return map[string]any{}, nil
	}

	return attributes, nil
}

// Merge current attributes stored in an authentik user so we don't lose any data
func updateUserAttributes(uid any, updates map[string]any) error {
	attrs, err := getUserRequest(uid)
	if err != nil {
		return err
	}

	for key, value := range updates {
		attrs[key] = value
	}

	return updateUserRequest(uid, map[string]any{
		"attributes": attrs,
	})
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
