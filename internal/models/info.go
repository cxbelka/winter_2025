package models

type InfoResponse struct {
	Balance   int
	Inventory []InventoryItem
	Transfers struct {
		Received []ReceivedTransfer
		Sent     []SentTransfer
	}
}

type InventoryItem struct {
	Type string
	Qty  int
}

type ReceivedTransfer struct {
	From   string
	Amount int
}

type SentTransfer struct {
	To     string
	Amount int
}
