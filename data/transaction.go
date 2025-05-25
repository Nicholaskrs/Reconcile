package data

import (
	"github.com/shopspring/decimal"
	"time"
)

type TransactionType string

const (
	TTDebit  TransactionType = "debit"
	TTCredit TransactionType = "credit"
)

type SystemTransaction struct {
	ID              string
	Amount          decimal.Decimal
	Type            TransactionType
	TransactionTime time.Time
}

type BankTransaction struct {
	ID              string
	Amount          decimal.Decimal
	TransactionDate time.Time
}
