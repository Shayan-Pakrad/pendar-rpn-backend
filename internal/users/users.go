package users

import (
	"net/http"

	"github.com/Shayan-Pakrad/rpn-web-service/internal/auth"
	"github.com/Shayan-Pakrad/rpn-web-service/internal/db"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Signup handler
func Signup(c echo.Context) error {
	type SignupRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	req := new(SignupRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to hash password")
	}

	_, err = db.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", req.Username, string(hashedPassword))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(http.StatusCreated, "User created")
}

// Login handler
func Login(c echo.Context) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	var hashedPassword string
	err := db.DB.QueryRow("SELECT password FROM users WHERE username = $1", req.Username).Scan(&hashedPassword)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid username or password")
	}

	token, err := auth.GenerateJWT(req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}
