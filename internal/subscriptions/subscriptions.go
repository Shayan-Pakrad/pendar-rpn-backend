package subscriptions

import (
	"net/http"
	"time"

	"github.com/Shayan-Pakrad/rpn-web-service/internal/db"

	"github.com/labstack/echo/v4"
)

func CreateOrRenew(c echo.Context) error {
	type SubscriptionRequest struct {
		Username           string `json:"username"`
		Remaining_requests int    `json:"remaining_requests"`
		ExpiryDate         string `json:"expiry_date"` // "YYYY-MM-DD"
	}

	req := new(SubscriptionRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}

	expiry, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid date format")
	}

	_, err = db.DB.Exec(`INSERT INTO subscriptions (username, remaining_requests, expiry_date)
        VALUES ($1, $2, $3) 
        ON CONFLICT (username) DO UPDATE 
        SET remaining_requests = $2, expiry_date = $3`, req.Username, req.Remaining_requests, expiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to create or renew subscription")
	}

	return c.JSON(http.StatusOK, "Subscription updated")
}
