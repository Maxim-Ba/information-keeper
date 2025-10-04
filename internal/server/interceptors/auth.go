package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenKeeper interface {
	ValidateTokenWithBlacklist(ctx context.Context, token string) error
	GetUserFromAcssToken(token string) (*dto.UserRepoDTO, error)
}

func AuthInterceptor(tokenService TokenKeeper) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Info(fmt.Sprintf("AuthInterceptor info: %v", info.FullMethod))

		if info.FullMethod == "/auth.Auth/Login" ||
			info.FullMethod == "/auth.Auth/Register" ||
			info.FullMethod == "/auth.Auth/RefreshToken" ||
			info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}
		tokens := md["authorization"]
		logger.Info(fmt.Sprintf("AuthInterceptor tokens: %v", tokens))
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token not provided")
		}
		token := strings.TrimPrefix(tokens[0], "Bearer ")
		if err := tokenService.ValidateTokenWithBlacklist(ctx, token); err != nil {
			logger.Error(fmt.Sprintf("AuthInterceptor failed to validate token: %v", err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user, err := tokenService.GetUserFromAcssToken(token)
		if err != nil {
			logger.Error(fmt.Sprintf("AuthInterceptor failed to get user from token: %v", err))

			return nil, status.Error(codes.Unauthenticated, "failed to get user from token")
		}

		ctx = context.WithValue(ctx, "userID", user.ID)

		return handler(ctx, req)

	}
}
