package interfaces

import (
	"github.com/shopspring/decimal"
	"time"
)

type Service interface {
	ReconcileTransaction(in *ReconcileTransactionIn) *ReconcileTransactionOut
}

type ReconcileTransactionIn struct {
	SystemTransactionCsvPath string
	StartDate                time.Time
	EndDate                  time.Time

	// Key is bankIdentifier and the value is bank csv path.
	BankSystemCsvPaths map[string]string
}

type ReconcileTransactionOut struct {
	Success  bool
	ErrorMsg string

	TotalTransactionProcessedCount int
	MatchedTransactionCount        int
	UnmatchedTransactionCount      int

	// SystemUnmatchedTransaction is list of ID of transaction that couldn't be found in bank statement.
	SystemUnmatchedTransaction []string

	// BankUnmatchedTransactionMap is list of UniqueIdentifier grouped by bank for that couldn't be found in system statement.
	// Key is Bank name and value is array of UniqueIdentifier.
	BankUnmatchedTransactionMap map[string][]string

	// TotalUnmatchedAmount is sum of absolute differences in amount between matched transactions.
	TotalUnmatchedAmount decimal.Decimal
}
