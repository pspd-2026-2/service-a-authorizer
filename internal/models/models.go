package models

type AuthorizationRequest struct {
	UserID		string		`json:"userId"`
	Amount		float64		`json:"amount"`
	CardNumber	string		`json:"carNumber"`
	IPAddress	string		`json:"ipAddress"`
}

type AuthorizationResponse struct {
	TransactionID		string		`json:"transactionId"`
	AuthCode			string		`json:"authCode,omitempty"`
	Status				string		`json:"status"`
	CardStatus			string		`json:"cardStatus"`
	ApprovedLimit		float64		`json:"approvedLimit"`
	RequestedAmount 	float64		`json:"requestedAmount"`
	LimitSufficient		bool		`json:"limitSufficient"`
	Message				string		`json:"message"`
}

type AuthorizeTransactionInput struct {
	Request AuthorizationRequest
}

type AuthorizeTransactionOutput struct {
	Response AuthorizationResponse
	Error    error
}