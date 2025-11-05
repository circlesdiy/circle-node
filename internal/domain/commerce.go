package domain

import "time"

// Order represents a payment order for tickets
type Order struct {
	ID              string    `json:"id"`
	BuyerProfileID  string    `json:"buyer_profile_id"`
	TotalAmount     float64   `json:"total_amount"`
	Currency        string    `json:"currency"`
	PaymentStatus   string    `json:"payment_status"` // pending, completed, failed, refunded
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IsCompleted checks if the order payment is completed
func (o *Order) IsCompleted() bool {
	return o.PaymentStatus == PaymentStatusCompleted
}

// IsPending checks if the order payment is pending
func (o *Order) IsPending() bool {
	return o.PaymentStatus == PaymentStatusPending
}

// IsFailed checks if the order payment failed
func (o *Order) IsFailed() bool {
	return o.PaymentStatus == PaymentStatusFailed
}

// IsRefunded checks if the order was refunded
func (o *Order) IsRefunded() bool {
	return o.PaymentStatus == PaymentStatusRefunded
}

// Payment status constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

// Ticket represents a ticket for an event
type Ticket struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	ProfileID    string    `json:"profile_id"`
	OrderID      *string   `json:"order_id,omitempty"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	TicketStatus string    `json:"ticket_status"` // valid, used, cancelled, expired
	QRCode       string    `json:"qr_code"`
	IssuedAt     time.Time `json:"issued_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsValid checks if the ticket is valid and not used/cancelled/expired
func (t *Ticket) IsValid() bool {
	return t.TicketStatus == TicketStatusValid
}

// IsUsed checks if the ticket has been used
func (t *Ticket) IsUsed() bool {
	return t.TicketStatus == TicketStatusUsed
}

// IsCancelled checks if the ticket was cancelled
func (t *Ticket) IsCancelled() bool {
	return t.TicketStatus == TicketStatusCancelled
}

// IsExpired checks if the ticket has expired
func (t *Ticket) IsExpired() bool {
	return t.TicketStatus == TicketStatusExpired
}

// IsFree checks if the ticket is free (no charge)
func (t *Ticket) IsFree() bool {
	return t.Price == 0
}

// Ticket status constants
const (
	TicketStatusValid     = "valid"
	TicketStatusUsed      = "used"
	TicketStatusCancelled = "cancelled"
	TicketStatusExpired   = "expired"
)
