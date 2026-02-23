package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

// Picks is a slice of uint8 that marshals to a JSON array of integers
// instead of base64 (which is the default for []byte/[]uint8).
type Picks []uint8

// MarshalJSON implements json.Marshaler.
func (p Picks) MarshalJSON() ([]byte, error) {
	ints := make([]int, len(p))
	for i, v := range p {
		ints[i] = int(v)
	}
	return json.Marshal(ints)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Picks) UnmarshalJSON(data []byte) error {
	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		return err
	}
	*p = make(Picks, len(ints))
	for i, v := range ints {
		if v < 0 || v > 255 {
			return fmt.Errorf("pick value %d out of uint8 range at index %d", v, i)
		}
		(*p)[i] = uint8(v) //nolint:gosec // bounds checked above
	}
	return nil
}

// Game represents a game in API responses.
type Game struct {
	ID        int64     `json:"id"`
	Picks     Picks     `json:"picks"`
	CreatedAt time.Time `json:"created_at"`
}

// GameListResponse is the response for listing games.
type GameListResponse struct {
	Games      []Game `json:"games"`
	NextCursor *int64 `json:"next_cursor,omitempty"`
}

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
