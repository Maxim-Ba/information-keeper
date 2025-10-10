package interceptors

import (
	"context"
	"fmt"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TokenKeeper interface {
	ValidateTokenWithBlacklist(ctx context.Context, token string) error
	GetUserFromAcssToken(token string) (*dto.UserRepoDTO, error)
	GetAccessTokenFromContext(ctx context.Context) (string, error)
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

		token, err := tokenService.GetAccessTokenFromContext(ctx)
		if err != nil {

			return nil, status.Error(codes.Unauthenticated, err.Error())
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
