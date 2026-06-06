package pb

type AuthorizationRequest struct {
	UserId     string
	Amount     float64
	CardNumber string
	IpAddress  string
}

type AuthorizationResponse struct {
	TransactionId   string
	AuthCode        string
	Status          string
	CardStatus      string
	ApprovedLimit   float64
	RequestedAmount float64
	LimitSufficient bool
	Message         string
}
