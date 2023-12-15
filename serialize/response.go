package serialize

import "github.com/mangohow/mangokit/errors"

type Response struct {
	Data  interface{}  `json:"data"`
	Error errors.Error `json:"error"`
}
