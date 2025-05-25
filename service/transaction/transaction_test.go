package transaction

import (
	"sort"
	"testing"
	"time"
	transactionInterface "transaction_checker/service/transaction/interfaces"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAlignmentCheckerAllMatch(t *testing.T) {
	svc := NewService()

	startDate, _ := time.Parse("2006-01-02", "2025-05-25")
	endDate := startDate

	out := svc.ReconcileTransaction(&transactionInterface.ReconcileTransactionIn{
		SystemTransactionCsvPath: "../../testdata/testcase-1/system.csv",
		BankSystemCsvPaths:       map[string]string{"BCA": "../../testdata/testcase-1/bank.csv"},
		StartDate:                startDate,
		EndDate:                  endDate,
	})

	assert.Equal(t, out.ErrorMsg, "")
	assert.True(t, out.Success)

	assert.Equal(t, 2, out.MatchedTransactionCount)
	assert.Equal(t, 0, out.UnmatchedTransactionCount)
	assert.Equal(t, 2, out.TotalTransactionProcessedCount)
	assert.Equal(t, make(map[string][]string), out.BankUnmatchedTransactionMap)
	assert.Equal(t, make([]string, 0), out.SystemUnmatchedTransaction)
	assert.True(t, out.TotalUnmatchedAmount.Equal(decimal.NewFromInt(0)))
}

// Test case:
// 43 valid-range transactions (25 for Bank A, 18 for Bank B, 7 mismatched) + 20 before + 5 after the range
// 25 matched, 1 mismatched, 3 extra
// 18 matched, 3 mismatched, 2 extra
// 43 is valid
func TestAlignmentCheckerRangeWithDiscrepancy(t *testing.T) {
	svc := NewService()

	startDate, _ := time.Parse("2006-01-02", "2025-05-25")
	endDate, _ := time.Parse("2006-01-02", "2025-05-30")

	out := svc.ReconcileTransaction(&transactionInterface.ReconcileTransactionIn{
		SystemTransactionCsvPath: "../../testdata/testcase-2/system.csv",
		BankSystemCsvPaths: map[string]string{
			"BCA": "../../testdata/testcase-2/bank_a.csv",
			"BCB": "../../testdata/testcase-2/bank_b.csv",
		},
		StartDate: startDate,
		EndDate:   endDate,
	})

	expectedUnmatchedTransactionIds := []string{
		0: "sys43",
		1: "sys49",
		2: "sys48",
		3: "sys40",
		4: "sys45",
		5: "sys44",
		6: "sys47",
	}
	sort.Strings(expectedUnmatchedTransactionIds)
	sort.Strings(out.SystemUnmatchedTransaction)

	expectedBankUnmachedTransaction := map[string][]string{
		"BCA": {
			0: "bankA_extra2",
			1: "bankA_extra1",
			2: "bankA_sys48",
			3: "bankA_extra0",
			4: "bankA_sys45",
		},
		"BCB": {
			0: "bankB_extra1",
			1: "bankB_sys43",
			2: "bankB_extra0",
			3: "bankB_sys40",
		},
	}

	for bankCode, bankUnmatchedTransaction := range out.BankUnmatchedTransactionMap {
		sort.Strings(bankUnmatchedTransaction)
		sort.Strings(expectedBankUnmachedTransaction[bankCode])
	}

	assert.Equal(t, out.ErrorMsg, "")
	assert.True(t, out.Success)
	assert.True(t, out.Success)
	assert.Equal(t, 43, out.MatchedTransactionCount)
	assert.Equal(t, 16, out.UnmatchedTransactionCount)
	assert.Equal(t, 59, out.TotalTransactionProcessedCount)
	assert.Equal(t, expectedUnmatchedTransactionIds, out.SystemUnmatchedTransaction)
	assert.Equal(t, expectedBankUnmachedTransaction, out.BankUnmatchedTransactionMap)
	assert.True(t, out.TotalUnmatchedAmount.Equal(decimal.NewFromFloat(1199.6)))
}

// Test case:
// 20 valid-range transactions
//   - 5 has same value (2 for bank A, 2 for bank B 1 missing)
//   - 5 missing
//   - 5 transaction has same value (5 for bank A)
//   - 5 transaction has same value (5 for bank B)
//   - 10 extra transaction on bank A & B with same value
func TestAlignmentCheckerWithSameAmountOnSameDate(t *testing.T) {
	svc := NewService()

	startDate, _ := time.Parse("2006-01-02", "2025-05-25")
	endDate, _ := time.Parse("2006-01-02", "2025-05-30")

	out := svc.ReconcileTransaction(&transactionInterface.ReconcileTransactionIn{
		SystemTransactionCsvPath: "../../testdata/testcase-3/system.csv",
		BankSystemCsvPaths: map[string]string{
			"BCA": "../../testdata/testcase-3/bank_a.csv",
			"BCB": "../../testdata/testcase-3/bank_b.csv",
		},
		StartDate: startDate,
		EndDate:   endDate,
	})

	expectedUnmatchedTransactionIds := []string{
		0: "sys_day1_shared0",
		1: "sys_day1_extra0",
		2: "sys_day1_extra1",
		3: "sys_day1_extra2",
		4: "sys_day1_extra3",
		5: "sys_day1_extra4",
	}
	sort.Strings(expectedUnmatchedTransactionIds)
	sort.Strings(out.SystemUnmatchedTransaction)

	expectedBankUnmachedTransaction := map[string][]string{
		"BCA": {
			0: "bankA_extra0",
			1: "bankA_extra1",
			2: "bankA_extra2",
			3: "bankA_extra3",
			4: "bankA_extra4",
		},
		"BCB": {
			0: "bankB_extra0",
			1: "bankB_extra1",
			2: "bankB_extra2",
			3: "bankB_extra3",
			4: "bankB_extra4",
		},
	}

	for bankCode, bankUnmatchedTransaction := range out.BankUnmatchedTransactionMap {
		sort.Strings(bankUnmatchedTransaction)
		sort.Strings(expectedBankUnmachedTransaction[bankCode])
	}

	assert.Equal(t, out.ErrorMsg, "")
	assert.True(t, out.Success)
	assert.True(t, out.Success)
	assert.Equal(t, 14, out.MatchedTransactionCount)
	assert.Equal(t, 16, out.UnmatchedTransactionCount)
	assert.Equal(t, 30, out.TotalTransactionProcessedCount)
	assert.Equal(t, expectedUnmatchedTransactionIds, out.SystemUnmatchedTransaction)
	assert.Equal(t, expectedBankUnmachedTransaction, out.BankUnmatchedTransactionMap)
	assert.True(t, out.TotalUnmatchedAmount.Equal(decimal.NewFromFloat(2468.71)))
}
