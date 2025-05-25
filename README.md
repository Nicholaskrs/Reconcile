# 🧾 Transaction Reconciliation Tool

This tool compares system transactions against one or more bank transaction files to identify mismatches in dates and amounts.

## 🚀 How to Run

### 1. Run the Main Program

To run the reconciliation tool using `main.go`:

```bash
go run main.go
```

This will execute the reconciliation logic using CSV paths defined in the `main.go` file.

> ✅ Make sure the input CSV files exist and follow the expected format.

### 2. Run via Test

Alternatively, you can run the reconciliation logic through test cases:

```bash
go test ./...
```

This will run all test files and validate reconciliation behavior, including edge cases and custom scenarios.

---

## 📂 CSV Format

### System Transaction CSV (4 columns):

| ID       | Amount   | Type    | TransactionTime         |
|----------|----------|---------|--------------------------|
| UUID     | 10000.00 | CREDIT  | 2025-05-25 10:00:00     |

- **Type** should be `CREDIT` or `DEBIT`
- **TransactionTime** must follow format `2006-01-02 15:04:05`

### Bank Transaction CSV (3 columns):

| ID       | Amount   | TransactionDate |
|----------|----------|-----------------|
| UUID     | -10000.00| 2025-05-25      |

- **Amount** for debits should be **negative**
- **TransactionDate** must follow format `2006-01-02`
