package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Maxim-Ba/information-keeper/pkg/proto"

	"google.golang.org/grpc/metadata"
)

// WithAuthToken добавляет токен в метаданные gRPC.
func WithAuthToken(ctx context.Context, token string) context.Context {
	md := metadata.Pairs("authorization", "bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}

// PrintArtifactInfo печатает информацию о артефакте.
func PrintArtifactInfo(artifact *proto.Artifact) {
	slog.Info(fmt.Sprintf("ID: %s", artifact.Id))
	slog.Info(fmt.Sprintf("Type: %s", artifact.Type))
	slog.Info(fmt.Sprintf("Meta: %s", artifact.MetaInfo))
	slog.Info(fmt.Sprintf("Link: %s", artifact.Link))
	slog.Info(fmt.Sprintf("Created: %d", artifact.CreatedAt))
	slog.Info(fmt.Sprintf("Expired: %d", artifact.ExpiredAt))
}
