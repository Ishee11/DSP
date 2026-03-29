package model

type BidResponse struct {
	RequestID string  `json:"request_id"`
	ImpID     string  `json:"imp_id"`
	Price     float64 `json:"price"`
	AdID      string  `json:"ad_id"`
}
