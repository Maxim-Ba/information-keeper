package interceptors

import (
	"context"
	"strings"

	"github.com/Maxim-Ba/information-keeper/internal/server/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(tokenService *services.TokenService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if info.FullMethod == "/auth.Auth/Login" ||
			info.FullMethod == "/auth.Auth/Register" ||
			info.FullMethod == "/auth.Auth/RefreshToken" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}
		tokens := md["authorization"]
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token not provided")
		}
		token := strings.TrimPrefix(tokens[0], "Bearer ")
		if err := tokenService.ValidateTokenWithBlacklist(token); err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user, err := tokenService.GetUserFromAcssToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "failed to get user from token")
		}

		ctx = context.WithValue(ctx, "userID", user.ID)

		return handler(ctx, req)

	}
}
