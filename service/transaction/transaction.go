package transaction

import (
	"errors"
	"github.com/shopspring/decimal"
	"time"
	"transaction_reconciler/data"
	transactionInterface "transaction_reconciler/service/transaction/interfaces"
	"transaction_reconciler/util"
)

var _ transactionInterface.Service = (*Service)(nil)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// ReconcileTransaction compares system transactions against multiple bank statements
// within a specified date range. It identifies matched and unmatched transactions,
// calculates total discrepancies in amounts, and returns a detailed reconciliation report.
// StartDate and End date must have no time.
func (s *Service) ReconcileTransaction(in *transactionInterface.ReconcileTransactionIn) *transactionInterface.ReconcileTransactionOut {
	resp := &transactionInterface.ReconcileTransactionOut{}

	if in.StartDate.IsZero() {
		resp.ErrorMsg = "start date is empty"
		return resp
	}
	if in.StartDate.Hour() != 0 || in.StartDate.Minute() != 0 || in.StartDate.Second() != 0 {
		resp.ErrorMsg = "start date is invalid"
		return resp
	}

	if in.EndDate.IsZero() {
		resp.ErrorMsg = "end date is empty"
		return resp
	}
	if in.EndDate.Hour() != 0 || in.EndDate.Minute() != 0 || in.EndDate.Second() != 0 {
		resp.ErrorMsg = "end date is invalid"
		return resp
	}

	if in.EndDate.Before(in.StartDate) {
		resp.ErrorMsg = "end date is before start date"
		return resp
	}
	if in.SystemTransactionCsvPath == "" {
		resp.ErrorMsg = "system transaction csv path is empty"
		return resp
	}

	if len(in.BankSystemCsvPaths) == 0 {
		resp.ErrorMsg = "system transaction bank system csv path is empty"
		return resp
	}

	// We can fetch both system and bank transaction on same times.
	resultsCh, errCh := util.ParseCSVRecordsAsync(in.SystemTransactionCsvPath, convertSystemTransactionRow)

	// bankDetailMap maps UUIDs to their corresponding BankTransaction.
	bankDetailMap := make(map[string]*data.BankTransaction)

	// Key is BankTransaction Date and Amount,
	// the value is array of BankTransaction's ID that have given Date and Amount.
	bankStatements := make(map[time.Time]map[string][]string)
	// key is UUID and value is bankUUID.
	bankUUIDMap := make(map[string]string)
	for bankUUID, bankSystemPath := range in.BankSystemCsvPaths {
		if bankSystemPath == "" {
			resp.ErrorMsg = "bank system path is empty"
			return resp
		}
		bankTransactions, err := util.ParseCSVRecords(bankSystemPath, convertBankTransactionRow)
		if err != nil {
			resp.ErrorMsg = err.Error()
			return resp
		}
		for _, bankTransaction := range bankTransactions {
			bankUUIDMap[bankTransaction.ID] = bankUUID
			bankDetailMap[bankTransaction.ID] = bankTransaction
			if bankStatements[bankTransaction.TransactionDate] == nil {
				bankStatements[bankTransaction.TransactionDate] = make(map[string][]string)
			}
			bankStatements[bankTransaction.TransactionDate][bankTransaction.Amount.String()] = append(
				bankStatements[bankTransaction.TransactionDate][bankTransaction.Amount.String()],
				bankTransaction.ID,
			)
		}
	}

	systemTransactions := make([]*data.SystemTransaction, 0)

	// Key is system ID and value is corresponding SystemTransaction.
	systemTransactionMap := make(map[string]*data.SystemTransaction)

	// Key is SystemTransaction Date (without time) and Amount,
	// the value is array of systemTransaction ID that have given Date and Amount.
	systemTransactionStatement := make(map[time.Time]map[string][]string)

	select {
	case results := <-resultsCh:
		systemTransactions = results
	case err := <-errCh:
		resp.ErrorMsg = err.Error()
		return resp
	}

	for _, systemTransaction := range systemTransactions {
		systemTransactionMap[systemTransaction.ID] = systemTransaction
		transactionTime := systemTransaction.TransactionTime
		transactionDate := time.Date(
			transactionTime.Year(), transactionTime.Month(), transactionTime.Day(),
			0, 0, 0, 0,
			transactionTime.Location(),
		)
		if systemTransactionStatement[transactionDate] == nil {
			systemTransactionStatement[transactionDate] = make(map[string][]string)
		}

		amount := systemTransaction.Amount

		// When SystemTransaction type debit, bank statement will record it as negative value.
		if systemTransaction.Type == data.TTDebit {
			amount = amount.Mul(decimal.NewFromInt(-1))
		}

		systemTransactionStatement[transactionDate][amount.String()] = append(
			systemTransactionStatement[transactionDate][amount.String()],
			systemTransaction.ID,
		)
	}

	matchedTransactionCount := 0
	unmatchedTransactionCount := 0

	// Key is BankUUID and value is bankTransaction's IDs.
	bankUnmatchedTransactionMap := make(map[string][]string)
	systemUnmatchedTransactionIds := make([]string, 0)
	totalUnmatchedAmount := decimal.NewFromInt(0)

	date := in.StartDate
	// Loop daily until the end date.
	for !date.After(in.EndDate) {

		// First we'll match system transaction to all bank statements.
		dailySystemTransactionList := systemTransactionStatement[date]
		for amount, systemTransactionIds := range dailySystemTransactionList {

			// Check if count of bank statement and system statement match.
			if len(bankStatements[date][amount]) > len(systemTransactionIds) {
				// There's missing statement on system

				// KNOWN ISSUE: on case multiple amount we cannot be sure which one of statement system is not paid.
				// Example : 10 march 2022 and there's 10 statement with 100k value, only 9 statement on system
				// we cannot be sure which of the bank statement is not recorded on system.

				unmatchedTransactionCount += len(bankStatements[date][amount]) - len(systemTransactionIds)
				matchedTransactionCount += len(systemTransactionIds)
				for i := 0; i < (len(bankStatements[date][amount]) - len(systemTransactionIds)); i++ {
					bankDetail := bankDetailMap[bankStatements[date][amount][i]]
					bankUUID := bankUUIDMap[bankDetail.ID]
					bankUnmatchedTransactionMap[bankUUID] = append(
						bankUnmatchedTransactionMap[bankUUID],
						bankDetail.ID,
					)
					totalUnmatchedAmount = totalUnmatchedAmount.Add(bankDetail.Amount)
				}
			} else if len(bankStatements[date][amount]) < len(systemTransactionIds) {
				// There's missing statement on bank.
				unmatchedTransactionCount += len(systemTransactionIds) - len(bankStatements[date][amount])
				matchedTransactionCount += len(bankStatements[date][amount])
				for i := 0; i < (len(systemTransactionIds) - len(bankStatements[date][amount])); i++ {
					systemStatement := systemTransactionMap[systemTransactionIds[i]]
					systemUnmatchedTransactionIds = append(systemUnmatchedTransactionIds, systemTransactionIds[i])
					totalUnmatchedAmount = totalUnmatchedAmount.Add(systemStatement.Amount)
				}
			} else {
				matchedTransactionCount += len(systemTransactionIds)
			}

			// Remove bank statement, we want to put all untapped amount to report later.
			bankStatements[date][amount] = nil
		}

		// If there's still statement on given date that haven't been emptied then it should be missing on system.
		if len(bankStatements[date]) != 0 {

			// missing system statement
			for _, bankStatementIds := range bankStatements[date] {
				unmatchedTransactionCount += len(bankStatementIds)
				for _, bankStatementId := range bankStatementIds {
					bankDetail := bankDetailMap[bankStatementId]
					bankUUID := bankUUIDMap[bankDetail.ID]
					bankUnmatchedTransactionMap[bankUUID] = append(
						bankUnmatchedTransactionMap[bankUUID],
						bankDetail.ID,
					)
					totalUnmatchedAmount = totalUnmatchedAmount.Add(bankDetail.Amount)
				}
			}
		}

		date = date.AddDate(0, 0, 1)
	}

	resp.Success = true
	resp.BankUnmatchedTransactionMap = bankUnmatchedTransactionMap
	resp.SystemUnmatchedTransaction = systemUnmatchedTransactionIds
	resp.UnmatchedTransactionCount = unmatchedTransactionCount
	resp.MatchedTransactionCount = matchedTransactionCount
	resp.TotalTransactionProcessedCount = matchedTransactionCount + unmatchedTransactionCount
	resp.TotalUnmatchedAmount = totalUnmatchedAmount
	return resp
}

// convertSystemTransactionRow parses a CSV row into a SystemTransaction.
// Expected format: ID, Amount, Type (DEBIT|CREDIT), Timestamp (2006-01-02 15:04:05)
func convertSystemTransactionRow(csvRow []string) (*data.SystemTransaction, error) {
	if len(csvRow) != 4 {
		return nil, errors.New("wrong number of fields in row")
	}
	amount, err := decimal.NewFromString(csvRow[1])
	if err != nil {
		return nil, err
	}
	transactionType := data.TransactionType(csvRow[2])

	validTransactionType := map[data.TransactionType]bool{
		data.TTDebit:  true,
		data.TTCredit: true,
	}
	if validTransactionType[transactionType] == false {
		return nil, errors.New("invalid transaction type")
	}

	// Layout must be exactly this reference time: "2006-01-02 15:04:05"
	transactionTime, err := time.Parse("2006-01-02 15:04:05", csvRow[3])
	if err != nil {
		return nil, err
	}

	return &data.SystemTransaction{
		ID:              csvRow[0],
		Amount:          amount,
		Type:            transactionType,
		TransactionTime: transactionTime,
	}, nil
}

// convertBankTransactionRow parses a CSV row into a BankTransaction.
// Expected format: ID, Amount, Date (2006-01-02)
func convertBankTransactionRow(csvRow []string) (*data.BankTransaction, error) {
	if len(csvRow) != 3 {
		return nil, errors.New("wrong number of fields in row")
	}
	amount, err := decimal.NewFromString(csvRow[1])
	if err != nil {
		return nil, err
	}

	// Layout must be exactly this reference time: "2006-01-02"
	transactionTime, err := time.Parse("2006-01-02", csvRow[2])
	if err != nil {
		return nil, err
	}

	return &data.BankTransaction{
		ID:              csvRow[0],
		Amount:          amount,
		TransactionDate: transactionTime,
	}, nil
}
