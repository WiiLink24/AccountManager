package tests

import (
	"encoding/json"
	"testing"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/stretchr/testify/assert"
)

// TestMiddlewareClaimsNoAttributes mocks the claims parser of the middleware as if the current user
// is brand new with no Wii's linked.
func TestMiddlewareClaimsNoAttributes(t *testing.T) {
	testJson := `{
		"email": "test@google.com",
		"name": "test",
		"preferred_username": "testuser",
		"groups": [],
		"wiis": [],
		"sub": "1",
		"public_profile": false
	}`

	var claims middleware.Claims
	err := json.Unmarshal([]byte(testJson), &claims)
	assert.NoError(t, err)

	assert.Equal(t, "test@google.com", claims.Email)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test", claims.Name)
	assert.Equal(t, "1", claims.UserId)
	assert.Empty(t, claims.Groups)
	assert.Empty(t, claims.Wiis)
	assert.False(t, claims.PublicProfile)
}

// TestMiddlewareClaimsHasAttributes mocks the claims parser of the middleware as if the current user
// has Wiis
func TestMiddlewareClaimsHasAttributes(t *testing.T) {
	testJson := `{
		"email": "test@google.com",
		"preferred_username": "testuser",
		"name": "test",
		"sub": "1",
		"groups": [],
		"wiis": [
			{
				"wii_number": "7052226925690371", 
				"hollywood_id": 610515068, 
				"serial_number": "", 
				"dominos_linked": true, 
				"just_eat_linked": false
			}
		], 
		"public_profile": false
	}`

	var claims middleware.Claims
	err := json.Unmarshal([]byte(testJson), &claims)
	assert.NoError(t, err)

	assert.Equal(t, "test@google.com", claims.Email)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test", claims.Name)
	assert.Equal(t, "1", claims.UserId)
	assert.Empty(t, claims.Groups)
	assert.Equal(t, false, claims.PublicProfile)

	assert.NotNil(t, claims.Wiis)
	wii := claims.Wiis[0]
	assert.Equal(t, "7052226925690371", wii.WiiNumber)
	assert.Equal(t, 610515068, wii.HollywoodID)
	assert.Equal(t, true, wii.DominosLinked)
	assert.Equal(t, false, wii.JustEatLinked)
}
