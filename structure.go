package main

import (
	"encoding/xml"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type JWTClaims struct {
	Username string `json:"nickname"`
	PFP      string `json:"picture"`
	Email    string
	jwt.RegisteredClaims
}

type OIDCConfig struct {
	XMLName      xml.Name `xml:"oidc"`
	ClientID     string   `xml:"clientID"`
	ClientSecret string   `xml:"clientSecret"`
	RedirectURL  string   `xml:"redirectURL"`
	Scopes       []string `xml:"scopes"`
	Provider     string   `xml:"provider"`
	LogoutURL    string   `xml:"logoutURL"`
}

type DiscordOAuthConfig struct {
	ClientID     string `xml:"clientID"`
	ClientSecret string `xml:"clientSecret"`
	RedirectURL  string `xml:"redirectURL"`
}

type Config struct {
	Username                string             `xml:"username"`
	Password                string             `xml:"password"`
	WiiMailUsername         string             `xml:"wiiMailUsername"`
	WiiMailPassword         string             `xml:"wiiMailPassword"`
	DatabaseAddress         string             `xml:"databaseAddress"`
	DatabaseName            string             `xml:"databaseName"`
	WiiMailDatabaseAddress  string             `xml:"wiiMailDatabaseAddress"`
	WiiMailDatabaseName     string             `xml:"wiiMailDatabaseName"`
	DominosDatabaseAddress  string             `xml:"dominosDatabaseAddress"`
	DominosDatabaseName     string             `xml:"dominosDatabaseName"`
	DominosDatabaseUsername string             `xml:"dominosDatabaseUsername"`
	DominosDatabasePassword string             `xml:"dominosDatabasePassword"`
	Address                 string             `xml:"address"`
	OIDCConfig              OIDCConfig         `xml:"oidc"`
	DiscordOAuthConfig      DiscordOAuthConfig `xml:"discord"`
}

type AppAuthConfig struct {
	OAuth2Config *oauth2.Config
	Provider     *oidc.Provider
}
