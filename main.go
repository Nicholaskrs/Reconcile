package main

import (
	"transaction_checker/service/transaction"
	"transaction_checker/service/transaction/interfaces"
)

func main() {

	transactionService := transaction.NewService()
	transactionService.ReconcileTransaction(&interfaces.ReconcileTransactionIn{})

}
