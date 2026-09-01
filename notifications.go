package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logrusorgru/aurora/v4"
)

var testHTTPClient = &http.Client{Timeout: 30 * time.Second}

type notificationCategory struct {
	Key   string
	Label string
}

var notificationCategories = []notificationCategory{
	{"nintendo_channel", "Nintendo Channel"},
	{"evc", "Everybody Votes Channel"},
	{"cmoc", "Check Mii Out Channel"},
	{"wii_room", "Wii Room"},
	{"general_announcements", "General Announcements"},
	{"critical_announcements", "Critical Announcements"},
	{"newsletter_announcements", "Newsletter Announcements"},
	{"food_channel", "Food Channel"},
	{"digicam_prints", "Digicam Prints"},
	{"kirby_tv_channel", "Kirby TV Channel"},
}

type categoryPageState struct {
	Key     string
	Label   string
	Enabled bool
}

func notificationState(username any) (masterEnabled bool, categories []categoryPageState, err error) {
	var count int
	if err = dbPool.QueryRow(ctx,
		"SELECT COUNT(*) FROM push_subscriptions WHERE username = $1", username).Scan(&count); err != nil {
		return false, nil, err
	}
	masterEnabled = count != 0

	// Missing preference rows mean the category is enabled.
	preferences := map[string]bool{}
	rows, err := dbPool.Query(ctx,
		"SELECT category, enabled FROM notification_preferences WHERE username = $1", username)
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		var enabled bool
		if err := rows.Scan(&category, &enabled); err != nil {
			return false, nil, err
		}
		preferences[category] = enabled
	}

	categories = make([]categoryPageState, 0, len(notificationCategories))
	for _, category := range notificationCategories {
		enabled, ok := preferences[category.Key]
		categories = append(categories, categoryPageState{
			Key:     category.Key,
			Label:   category.Label,
			Enabled: !ok || enabled,
		})
	}
	return masterEnabled, categories, nil
}

func NotificationsPage(c *gin.Context) {
	username, _ := c.Get("username")
	email, _ := c.Get("email")

	masterEnabled, categories, err := notificationState(username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	data := gin.H{
		"username":      username,
		"email":         email,
		"notifications": masterEnabled,
		"categories":    categories,
		"checkout_url":  strings.TrimSuffix(config.Notifications.CheckoutURL, "/"),
	}
	if pfp, ok := c.Get("picture"); ok {
		data["pfp"] = pfp
	}
	c.HTML(http.StatusOK, "notifications.html", data)
}

func updateNotificationPreference(c *gin.Context) {
	username, _ := c.Get("username")

	var body struct {
		Category string `json:"category"`
		Enabled  *bool  `json:"enabled"`
	}
	if err := c.BindJSON(&body); err != nil || body.Category == "" || body.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "category and enabled are required",
		})
		return
	}

	valid := false
	for _, category := range notificationCategories {
		if category.Key == body.Category {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "unknown category",
		})
		return
	}

	_, err := dbPool.Exec(ctx, `
		INSERT INTO notification_preferences (username, category, enabled)
		VALUES ($1, $2, $3)
		ON CONFLICT (username, category) DO UPDATE SET enabled = EXCLUDED.enabled
	`, username, body.Category, *body.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func notificationStatus(c *gin.Context) {
	username, _ := c.Get("username")

	var count int
	err := dbPool.QueryRow(ctx,
		"SELECT COUNT(*) FROM push_subscriptions WHERE username = $1", username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"enabled": count != 0,
		"count":   count,
	})
}

func sendTestNotification(c *gin.Context) {
	username, _ := c.Get("username")

	body, err := json.Marshal(map[string]string{"username": username.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	request, err := http.NewRequest(
		"POST",
		strings.TrimSuffix(config.Notifications.CheckoutURL, "/")+"/notifications/test",
		bytes.NewReader(body),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	request.Header.Set("Authorization", "Bearer "+config.Notifications.SharedSecret)
	request.Header.Set("Content-Type", "application/json")

	response, err := testHTTPClient.Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Checkout refused to comment :( %v", err),
		})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(aurora.Red("error closing body:"), err)
		}
	}(response.Body)

	// Relay Checkout's answer if error
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "failed to read Checkout response",
		})
		return
	}
	c.Data(response.StatusCode, "application/json", responseBody)
}
