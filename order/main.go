package order

type Order struct {
    Symbol        string  `json:"symbol"`
    OrderID       int64   `json:"orderId"`
    Price         string  `json:"price"`
    Quantity      string  `json:"quantity"`
    Type          string  `json:"type"`
    Status        string  `json:"status"`
    Side          string  `json:"side"`
    Time          int64   `json:"time"`
}