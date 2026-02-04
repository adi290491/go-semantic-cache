package model

import "time"

type ResponseModel struct {
	Response   string
	Query      string
	Embedding  []byte
	Created_at time.Time
}
