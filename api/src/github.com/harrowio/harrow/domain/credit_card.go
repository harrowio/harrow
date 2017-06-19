package domain

/**
 *
 *  CAUTION!
 *
 *  This type is used withing billing events as well; only make
 *  backwards compatible hchanges with regards to serialization:
 *
 *  - if you rename a field, keep the old name in the "json" tag
 *  - do not remove fields, only add new fields
 *
 **/
type CreditCard struct {
	// IsDefault is set to true if the payment provider reports this
	// credit card as the default mode of payment.
	IsDefault bool `json:"isDefault"`

	CardholderName string `json:"cardholderName"`

	// SafeCardNumber is a truncated version of the credit card
	// number, safe for display purposes.
	SafeCardNumber string `json:"safeCardNumber"`

	// CardId identifies this card at the payment provider
	CardId string `json:"cardId"`

	// Token is a provider specific identifier for using this
	// credit card in transactions.
	Token string `json:"token"`
}

func NewCreditCard(cardId string) *CreditCard {
	return &CreditCard{
		CardId: cardId,
	}
}
