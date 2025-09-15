package riot

import "fmt"

type RiotAPIError struct {
	StatusCode int
	Message    string
	RetryAfter string
}

func (e *RiotAPIError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("Riot API Error (status: %d, retry after: %s): %s",
			e.StatusCode, e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("Riot API Error (status: %d): %s", e.StatusCode, e.Message)
}

func (e *RiotAPIError) IsRateLimit() bool {
	return e.StatusCode == 429
}

func (e *RiotAPIError) IsTemporary() bool {
	return e.StatusCode >= 500 || e.StatusCode == 429
}
