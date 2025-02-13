package models

type InfoResponse struct {
	Balance   int             `json:"coins"`
	Inventory []InventoryItem `json:"inventory"`
	Transfers struct {
		Received []ReceivedTransfer `json:"received"`
		Sent     []SentTransfer     `json:"sent"`
	} `json:"coinHistory"`
}

type InventoryItem struct {
	Type string `json:"type"`
	Qty  int    `json:"quantity"`
}

type ReceivedTransfer struct {
	From   string `json:"fromUser"`
	Amount int    `json:"amount"`
}

type SentTransfer struct {
	To     string `json:"toUser"`
	Amount int    `json:"amount"`
}
