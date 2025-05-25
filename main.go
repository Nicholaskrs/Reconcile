package main

import (
	"fmt"
	"time"
	"transaction_reconciler/service/transaction"
	"transaction_reconciler/service/transaction/interfaces"
)

func main() {

	transactionService := transaction.NewService()

	startDate, _ := time.Parse("2006-01-02", "2025-05-25")
	endDate, _ := time.Parse("2006-01-02", "2025-05-30")

	result := transactionService.ReconcileTransaction(&interfaces.ReconcileTransactionIn{
		SystemTransactionCsvPath: "testdata/testcase-2/system.csv",
		StartDate:                startDate,
		EndDate:                  endDate,
		BankSystemCsvPaths: map[string]string{
			"BCA": "testdata/testcase-2/bank_a.csv",
			"BCB": "testdata/testcase-2/bank_b.csv",
		},
	})
	PrintReconcileResult(result)

}

func PrintReconcileResult(out *interfaces.ReconcileTransactionOut) {
	if out == nil {
		fmt.Println("No result to display.")
		return
	}

	if !out.Success {
		fmt.Printf("❌ Reconciliation failed: %s\n", out.ErrorMsg)
		return
	}

	fmt.Println("✅ Reconciliation Summary")
	fmt.Println("------------------------------")
	fmt.Printf("Total Processed Transactions : %d\n", out.TotalTransactionProcessedCount)
	fmt.Printf("Matched Transactions         : %d\n", out.MatchedTransactionCount)
	fmt.Printf("Unmatched Transactions       : %d\n", out.UnmatchedTransactionCount)
	fmt.Printf("Total Unmatched Amount       : %s\n", out.TotalUnmatchedAmount.String())
	fmt.Println()

	// System unmatched transactions
	if len(out.SystemUnmatchedTransaction) > 0 {
		fmt.Println("📌 System Unmatched Transactions:")
		for _, id := range out.SystemUnmatchedTransaction {
			fmt.Printf("  - %s\n", id)
		}
		fmt.Println()
	}

	// Bank unmatched transactions grouped by bank
	if len(out.BankUnmatchedTransactionMap) > 0 {
		fmt.Println("🏦 Bank Unmatched Transactions:")
		for bank, ids := range out.BankUnmatchedTransactionMap {
			fmt.Printf("  Bank: %s\n", bank)
			for _, id := range ids {
				fmt.Printf("    - %s\n", id)
			}
		}
	}
}
