package model

import "time"

type ResponseModel struct {
	Response   string    `redis:"response"`
	Query      string    `redis:"query"`
	Embedding  []byte    `redis:"embedding"`
	Created_at time.Time `redis:"created_at"`
}

type QueryRequestModel struct {
	Query string `json:"query"`
}
