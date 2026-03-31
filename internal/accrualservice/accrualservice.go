package accrualservice

import (
	"encoding/json"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/models"
	"io"
	"net/http"
	"strconv"
	"time"
)

type RetryAfterError struct {
	After time.Duration
}

func (e *RetryAfterError) Error() string {
	return "rate limited"
}

func Accrual(orderNumber string, url string, client *http.Client) (*models.Accrual, error) {

	var ans models.Accrual

	resp, err := client.Get(url + "/api/orders/" + orderNumber)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		dec := json.NewDecoder(resp.Body)

		err = dec.Decode(&ans)
		if err != nil {
			logger.Initialize().Info(err)
			return nil, err
		}
		return &ans, nil
	case http.StatusNoContent:
		// order is not registered in accrual system yet
		ans.Order = orderNumber
		ans.Status = "REGISTERED"
		ans.Accrual = nil
		_, _ = io.Copy(io.Discard, resp.Body)
		return &ans, nil
	case http.StatusTooManyRequests:
		ra := resp.Header.Get("Retry-After")
		sec, convErr := strconv.Atoi(ra)
		if convErr != nil || sec <= 0 {
			sec = 1
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, &RetryAfterError{After: time.Duration(sec) * time.Second}
	default:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, &RetryAfterError{After: 1 * time.Second}
	}

}
