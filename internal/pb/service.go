package pb

import "context"

type AuthorizationServiceServer interface {
	Authorize(context.Context, *AuthorizationRequest) (*AuthorizationResponse, error)
}

type AuthorizationServiceClient interface {
	Authorize(ctx context.Context, in *AuthorizationRequest) (*AuthorizationResponse, error)
}
