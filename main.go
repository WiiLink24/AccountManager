package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var (
	ctx        = context.Background()
	authConfig *AppAuthConfig
	config     Config
	verifier   *oidc.IDTokenVerifier
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("WiiLink Account Manager has encountered a fatal error! Reason: %v\n", err)
	}
}

func main() {
	config = GetConfig()

	provider, err := oidc.NewProvider(ctx, config.OIDCConfig.Provider)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	authConfig = &AppAuthConfig{
		OAuth2Config: &oauth2.Config{
			ClientID:     config.OIDCConfig.ClientID,
			ClientSecret: config.OIDCConfig.ClientSecret,
			RedirectURL:  config.OIDCConfig.RedirectURL,
			Scopes:       config.OIDCConfig.Scopes,
			Endpoint:     provider.Endpoint(),
		},
		Provider: provider,
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: config.OIDCConfig.ClientID})
	r := gin.Default()

	// Serve static files in debug mode
	if gin.Mode() == gin.DebugMode {
		r.Static("/assets", "./assets")
	}

	// Load HTML templates from the templates directory
	r.LoadHTMLGlob("templates/*")

	// Define routes and their handlers
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/login")
	})
	r.GET("/login", LoginPage)
	r.GET("/start", StartPanelHandler)
	r.GET("/authorize", FinishPanelHandler)

	auth := r.Group("/")
	auth.Use(middleware.AuthenticationMiddleware(verifier))
	{
		auth.GET("/manage", HomePage)
		auth.GET("/notlinked", NotLinkedPage)
		auth.POST("/dominos/link", linkDominos)
		auth.GET("/logout", logout)
		auth.GET("/refresh", refresh)
	}

	// Routes for API
	api := r.Group("/link")
	api.GET("/", linkRedirect)
	api.Use(middleware.AuthenticationPOSTMiddleware(verifier))
	{
		api.POST("/link", link)
		api.GET("/user/:uid", getUser)
	}

	// Start the server
	fmt.Printf("Starting HTTP connection (%s)...\nNot using the usual port for HTTP?\nBe sure to use a proxy, otherwise the Wii can't connect!\n", config.Address)
	log.Fatalln(r.Run(config.Address))
}
