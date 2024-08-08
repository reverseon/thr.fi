package commons

import "encoding/json"

type TypeReturnedURLInfo struct {
	Original_url       string
	Backhalf           string
	Password_protected bool
}

type TypeStoredURLInfo struct {
	Original_url string
	Password     *string // nil if not password protected
	Creator_id   *string // nil if not created by a user
}

func (s TypeStoredURLInfo) MarshalBinary() ([]byte, error) {
	// marshal as JSON
	return json.Marshal(s)
}

func (s *TypeStoredURLInfo) UnmarshalBinary(data []byte) error {
	// unmarshal from JSON
	return json.Unmarshal(data, s)
}
