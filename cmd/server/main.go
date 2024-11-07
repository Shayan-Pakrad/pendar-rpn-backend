package main

import (
	"log"

	"github.com/Shayan-Pakrad/rpn-web-service/config"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/auth"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/db"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/rpn"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/subscriptions"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/users"

	"github.com/labstack/echo/v4"
)

// Comment

func main() {
	// Initialize the database connection
	db.InitDB()

	defer func() {
		if err := db.DB.Close(); err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize private and public keys
	auth.InitKeys()

	// Start Echo
	e := echo.New()

	// User endpoints
	e.POST("/signup", users.Signup)
	e.POST("/login", users.Login)

	// rpn endpoints
	r := e.Group("/rpn")
	r.Use(auth.VerifyJWT, auth.CheckSubsctiption)
	r.POST("/evaluate", rpn.EvaluateExpression)

	// Admin endpoints
	a := e.Group("/admin")
	a.Use(auth.AdminMiddleware(cfg.AdminUsername, cfg.AdminPassword))
	a.POST("/subscriptions", subscriptions.CreateOrRenew)

	port := "8080"
	log.Fatal(e.Start(":" + port))
}
