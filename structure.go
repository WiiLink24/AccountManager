package main

import (
	"encoding/xml"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	XMLName             xml.Name `xml:"oidc"`
	ClientID            string   `xml:"clientID"`
	ClientSecret        string   `xml:"clientSecret"`
	RedirectURL         string   `xml:"redirectURL"`
	Scopes              []string `xml:"scopes"`
	Provider            string   `xml:"provider"`
	LogoutURL           string   `xml:"logoutURL"`
	ServiceAccountToken string   `xml:"serviceAccountToken"`
}

type Config struct {
	Address    string     `xml:"address"`
	OIDCConfig OIDCConfig `xml:"oidc"`
}

type AppAuthConfig struct {
	OAuth2Config *oauth2.Config
	Provider     *oidc.Provider
}
