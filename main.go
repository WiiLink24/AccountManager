package main

import (
	"context"
	"fmt"
	"log"

	"github.com/WiiLink24/AccountManager/middleware"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

var (
	ctx         = context.Background()
	pool        *pgxpool.Pool
	wiiMailPool *pgxpool.Pool
	authConfig  *AppAuthConfig
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("WiiLink Account Manager has encountered a fatal error! Reason: %v\n", err)
	}
}

func main() {
	config := GetConfig()

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

	// Connect account manager database
	dbString := fmt.Sprintf("postgres://%s:%s@%s/%s", config.Username, config.Password, config.DatabaseAddress, config.DatabaseName)
	pool, err = pgxpool.New(ctx, dbString)
	checkError(err)

	// Connect Wii Mail database
	dbString = fmt.Sprintf("postgres://%s:%s@%s/%s", config.WiiMailUsername, config.WiiMailPassword, config.WiiMailDatabaseAddress, config.WiiMailDatabaseName)
	wiiMailPool, err = pgxpool.New(ctx, dbString)
	checkError(err)

	defer pool.Close()
	defer wiiMailPool.Close()

	r := gin.Default()

	// Serve static files in debug mode
	if gin.Mode() == gin.DebugMode {
		r.Static("/assets", "./assets")
	}

	// Load HTML templates from the templates directory
	r.LoadHTMLGlob("templates/*")

	// Define routes and their handlers
	r.GET("/login", LoginPage)
	r.GET("/start", StartPanelHandler)
	r.GET("/authorize", FinishPanelHandler)

	auth := r.Group("/")
	auth.Use(middleware.AuthenticationMiddleware())
	{
		auth.GET("/manage", HomePage)
		auth.GET("/notlinked", NotLinkedPage)
		auth.GET("/link", LinkHandler)
	}

	// Start the server
	fmt.Printf("Starting HTTP connection (%s)...\nNot using the usual port for HTTP?\nBe sure to use a proxy, otherwise the Wii can't connect!\n", config.Address)
	log.Fatalln(r.Run(config.Address))
}
