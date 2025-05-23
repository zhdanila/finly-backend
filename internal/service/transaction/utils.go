package transaction

import (
	"finly-backend/internal/domain"
	"finly-backend/internal/domain/enums/e_transaction_type"
)

// invertDelta calculates the signed value of a transaction based on its type.
// If invert is true, the sign of the delta is reversed.
func invertDelta(transactionType string, amount float64, invert bool) (float64, error) {
	switch transactionType {
	case e_transaction_type.Deposit.String():
		if invert {
			return amount, nil
		}
		return -amount, nil
	case e_transaction_type.Withdrawal.String():
		if invert {
			return -amount, nil
		}
		return amount, nil
	default:
		return 0, errs.InvalidTransactionType
	}
}

// calculateNewAmount returns the new balance after applying the transaction.
// It takes into account the transaction type and current budget balance.
func calculateNewAmount(budgetHistory *domain.BudgetHistory, amount float64, transactionType string) (float64, error) {
	var newAmount float64
	switch transactionType {
	case e_transaction_type.Deposit.String():
		if budgetHistory == nil {
			newAmount = amount
		} else {
			newAmount = budgetHistory.Balance + amount
		}
	case e_transaction_type.Withdrawal.String():
		if budgetHistory == nil || budgetHistory.Balance < amount {
			return 0, errs.InsufficientBalance
		}
		newAmount = budgetHistory.Balance - amount
	default:
		return 0, errs.InvalidTransactionType
	}
	return newAmount, nil
}

// calculateDelta returns the balance delta for a transaction based on its type.
// Deposit increases balance, Withdrawal decreases it.
func calculateDelta(transactionType string, amount float64) (float64, error) {
	switch transactionType {
	case e_transaction_type.Deposit.String():
		return amount, nil
	case e_transaction_type.Withdrawal.String():
		return -amount, nil
	default:
		return 0, errs.InvalidTransactionType
	}
}

// calculateDeltaChange returns the net balance difference caused by changing a transaction.
// Useful when updating a transaction from one type/amount to another.
func calculateDeltaChange(oldType string, oldAmount float64, newType string, newAmount float64) (float64, error) {
	oldDelta, err := calculateDelta(oldType, oldAmount)
	if err != nil {
		return 0, err
	}

	newDelta, err := calculateDelta(newType, newAmount)
	if err != nil {
		return 0, err
	}

	return newDelta - oldDelta, nil
}
