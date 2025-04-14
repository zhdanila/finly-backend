package e_transaction_type

type Enum string

const (
	Deposit    Enum = "deposit"
	Withdrawal Enum = "withdrawal"
	Initial    Enum = "initial"
)

func (r *Enum) IsValid() bool {
	switch *r {
	case Deposit, Withdrawal:
		return true
	}
	return false
}

func (r Enum) String() string {
	return string(r)
}
