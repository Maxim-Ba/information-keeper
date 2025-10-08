package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Maxim-Ba/information-keeper/internal/domain"
	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/internal/server/interceptors"
	"github.com/Maxim-Ba/information-keeper/internal/server/services"
	eventidgen "github.com/Maxim-Ba/information-keeper/pkg/event-id-gen"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	pb "github.com/Maxim-Ba/information-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authService  AuthServiceInterface
	tokenService TokenKeeper

	pb.UnimplementedAuthServer
}

type ArtifactServer struct {
	artifactService ArtifactRW
	tokenService    TokenKeeper
	syncManager     Synchronizer

	pb.UnimplementedArtifactServiceServer
}
type HealthServer struct {
	pb.UnimplementedHealthServer
}
type PasswordServiceInterface interface {
	ChangePassword(ctx context.Context, data *dto.ChangePassword) error
}
type Synchronizer interface {
	Subscribe(userID string, clientID string) chan *pb.SyncEvent
	Unsubscribe(userID string, clientID string)
	GetEventsSince(userID string, since int64) []*pb.SyncEvent
}
type AuthServiceInterface interface {
	Login(ctx context.Context, login, password string) (*services.JWTToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*services.JWTToken, error)
	Logout(ctx context.Context, token *services.JWTToken) error
	Register(ctx context.Context, login, email, password string) (*services.JWTToken, error)
	SendEmailConfirmation(ctx context.Context, email string) error
	PasswordServiceInterface
}
type TokenKeeper interface {
	ValidateTokenWithBlacklist(ctx context.Context, token string) error
	GetAccessTokenFromContext(ctx context.Context) (string, error)
	GetUserFromAcssToken(token string) (*dto.UserRepoDTO, error)
}

type ArtifactReader interface {
	GetArtifacts(ctx context.Context, userID string, page int32, onpage int32) ([]domain.Artifact, error)
	GetArtifactByID(ctx context.Context, userID string, id string) (*domain.Artifact, error)
}
type ArtifactWriter interface {
CreateArtifact(ctx context.Context, userID string, req *proto.CreateArtifactRequest) error
	UpdateArtifact(ctx context.Context, userID string, artifact *domain.Artifact) error
	DeleteArtifact(ctx context.Context, userID string, id string) error
}
type ArtifactRW interface {
	ArtifactReader
	ArtifactWriter
}
type GRPCServer struct {
	server *grpc.Server
}

func NewGRPCServer(authService AuthServiceInterface, artifactService ArtifactRW, tokenService TokenKeeper, syncManager Synchronizer) *GRPCServer {
	authInterceptor := interceptors.AuthInterceptor(tokenService)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))

	authServer := &AuthServer{authService: authService, tokenService: tokenService}
	artifactServer := &ArtifactServer{artifactService: artifactService, tokenService: tokenService, syncManager: syncManager}
	healthServer := health.NewServer()

	pb.RegisterAuthServer(grpcServer, authServer)
	pb.RegisterArtifactServiceServer(grpcServer, artifactServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	healthServer.SetServingStatus("auth", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("artifact", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	return &GRPCServer{
		server: grpcServer,
	}
}

func (s *GRPCServer) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	slog.Info("HealthServer Check")
	return &pb.HealthCheckResponse{
		Status: "SERVING",
	}, nil
}

// Реализация методов Auth сервиса
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	logger.Info("AuthServer Login")
	jwt, err := s.authService.Login(ctx, req.User.Login, req.User.Password)
	if err != nil {
		
			logger.Error("Ошибка при авторизации",
				slog.String("error", err.Error()),
				slog.String("login", req.User.Login),
			)

			// Определяем appropriate gRPC код ошибки
			var grpcCode codes.Code
			switch {
			case errors.Is(err, services.ErrInvalidCredentials):
				grpcCode = codes.Unauthenticated
			case errors.Is(err, services.ErrUserNotFound):
				grpcCode = codes.NotFound

			default:
				grpcCode = codes.Internal
			}

			return nil, status.Errorf(grpcCode, "AuthService Login: %v", err)
		
	}
	return &pb.LoginUserResponse{
		RefreshToken: jwt.RefreshToken,
		AccessToken:  jwt.AcssToken,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginUserResponse, error) {
	// TODO: Реализовать обновление токена
	logger.Info("AuthServer RefreshToken")
	return &pb.LoginUserResponse{
		RefreshToken: "new_refresh_token",
		AccessToken:  "new_access_token",
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	slog.Info("AuthServer Logout")
	err := s.authService.Logout(ctx, &services.JWTToken{AcssToken: req.AccessToken, RefreshToken: req.RefreshToken})
	if err != nil {
		logger.Error("Ошибка при выходе из системы", slog.String("error", err.Error()))
		return nil, err
	}
	return &pb.LogoutResponse{}, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegistrationUserRequest) (*pb.RegistrationUserResponse, error) {
	slog.Info("AuthServer Register")
	_, err := s.authService.Register(ctx, req.User.Login, req.User.Email, req.User.Password)
	if err != nil {
		logger.Error("Ошибка при регистрации", slog.String("error", err.Error()))
		return nil, err
	}
	return &pb.RegistrationUserResponse{}, nil
}

func (s *AuthServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// TODO: Реализовать смену пароля
	slog.Info("AuthServer ChangePassword")

	return &pb.ChangePasswordResponse{}, nil
}

func (s *AuthServer) RestorePassword(ctx context.Context, req *pb.RestorePasswordRequest) (*pb.RestorePasswordResponse, error) {
	// TODO: Реализовать восстановление пароля
	slog.Info("AuthServer RestorePassword")

	return &pb.RestorePasswordResponse{}, nil
}

func (s *AuthServer) SendEmailConfirmation(ctx context.Context, req *pb.SendEmailConfirmationRequest) (*pb.SendEmailConfirmationResponse, error) {
	// TODO: Реализовать отправку подтверждения email
	slog.Info("AuthServer SendEmailConfirmation")

	return &pb.SendEmailConfirmationResponse{}, nil
}

// Реализация методов ArtifactService
func (s *ArtifactServer) CreateArtifact(ctx context.Context, req *pb.CreateArtifactRequest) (*pb.CreateArtifactResponse, error) {

	slog.Info("AuthServer CreateArtifact")
	user, err := getUserFomCtx(ctx, s.tokenService)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get user from token: %v", err))
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = s.artifactService.CreateArtifact(ctx, user.ID, req)
	if err != nil {
		logger.Error(fmt.Sprintf("ArtifactServer CreateArtifact Failed to create artifact: %v", err))
		return nil, status.Error(codes.Internal, "failed to create artifact")
	}
	artifact := &pb.Artifact{}

	return &pb.CreateArtifactResponse{
		Artifact: artifact,
	}, nil
}

func (s *ArtifactServer) GetArtifact(ctx context.Context, req *pb.GetArtifactRequest) (*pb.GetArtifactResponse, error) {
	// TODO: Реализовать получение артефакта
	slog.Info("AuthServer GetArtifact")

	return &pb.GetArtifactResponse{}, nil
}

func (s *ArtifactServer) ListArtifacts(ctx context.Context, req *pb.ListArtifactsRequest) (*pb.ListArtifactsResponse, error) {
	logger.Info("AuthServer ListArtifacts")
	user, err := getUserFomCtx(ctx, s.tokenService)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get user from token: %v", err))
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	artifacts, err := s.artifactService.GetArtifacts(ctx, user.ID, req.Page, req.OnPage)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get artifacts: %v", err))
		return nil, status.Error(codes.Internal, "failed to get artifacts")
	}
	pbArtifacts := make([]*pb.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		pbArtifact := domain.DomainToPBArtifact(&artifact)
		pbArtifacts = append(pbArtifacts, pbArtifact)
	}

	return &pb.ListArtifactsResponse{
		Artifacts: pbArtifacts,
	}, nil
}

func (s *ArtifactServer) UpdateArtifact(ctx context.Context, req *pb.UpdateArtifactRequest) (*pb.UpdateArtifactResponse, error) {
	// TODO: Реализовать обновление артефакта
	slog.Info("AuthServer UpdateArtifact")

	return &pb.UpdateArtifactResponse{}, nil
}

func (s *ArtifactServer) DeleteArtifact(ctx context.Context, req *pb.DeleteArtifactRequest) (*pb.DeleteArtifactResponse, error) {
	slog.Info("AuthServer DeleteArtifact")
	user, err := getUserFomCtx(ctx, s.tokenService)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get user from token: %v", err))
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	err = s.artifactService.DeleteArtifact(ctx, user.ID, req.ArtifactId)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get artifacts: %v", err))
		return nil, status.Error(codes.Internal, "failed to delete artifact")
	}
	return &pb.DeleteArtifactResponse{}, nil
}

func (s *ArtifactServer) GetWithOTP(ctx context.Context, req *pb.GetWithOTPRequest) (*pb.GetWithOTPResponse, error) {
	// TODO: Реализовать получение с OTP
	slog.Info("AuthServer GetWithOTP")

	return &pb.GetWithOTPResponse{}, nil
}

func (s *ArtifactServer) Sync(stream pb.ArtifactService_SyncServer) error {
	ctx := stream.Context()

	// Получаем пользователя из контекста
	user, err := getUserFomCtx(ctx, s.tokenService)
	if err != nil {
		return status.Error(codes.Unauthenticated, "failed to get user")
	}

	// Ждем первый запрос от клиента
	firstReq, err := stream.Recv()
	if err != nil {
		return err
	}

	clientID := firstReq.GetClientId()
	lastSyncTime := firstReq.GetLastSyncTime()

	logger.Info("Client connected to sync stream",
		"userID", user.ID,
		"clientID", clientID,
		"lastSyncTime", lastSyncTime)

	// Подписываемся на события
	syncCh := s.syncManager.Subscribe(user.ID, clientID)
	defer s.syncManager.Unsubscribe(user.ID, clientID)

	// Отправляем историю событий (если нужно)
	if lastSyncTime > 0 {
		recentEvents := s.syncManager.GetEventsSince(user.ID, lastSyncTime)
		for _, event := range recentEvents {
			if err := stream.Send(event); err != nil {
				return err
			}
		}

		// Отправляем событие завершения начальной синхронизации
		syncComplete := &pb.SyncEvent{
			EventId:   eventidgen.GenerateEventID(),
			Timestamp: time.Now().Unix(),
			EventType: &pb.SyncEvent_SyncComplete{
				SyncComplete: &pb.SyncCompleteEvent{
					SyncTime:    time.Now().Unix(),
					EventsCount: int32(len(recentEvents)),
				},
			},
		}
		if err := stream.Send(syncComplete); err != nil {
			return err
		}
	}

	// Канал для обработки входящих запросов от клиента
	requestCh := make(chan *pb.SyncRequest, 10)

	// Горутина для чтения запросов от клиента
	go func() {
		defer close(requestCh)
		for {
			req, err := stream.Recv()
			if err != nil {
				slog.Error("Error receiving from stream", "error", err)
				return
			}
			requestCh <- req
		}
	}()

	// Главный цикл обработки
	for {
		select {
		case <-ctx.Done():
			slog.Info("Sync stream context done", "clientID", clientID)
			return nil

		case req, ok := <-requestCh:
			if !ok {
				return nil // Канал закрыт
			}
			// Обрабатываем запросы от клиента (heartbeat и т.д.)
			slog.Debug("Received sync request", "clientID", req.ClientId, "lastSyncTime", req.LastSyncTime)

		case event, ok := <-syncCh:
			if !ok {
				slog.Info("Sync channel closed", "clientID", clientID)
				return nil
			}
			// Отправляем событие клиенту
			if err := stream.Send(event); err != nil {
				slog.Error("Error sending event to client", "error", err, "clientID", clientID)
				return err
			}
		}
	}
}
func getUserFomCtx(ctx context.Context, tokenService TokenKeeper) (*dto.UserRepoDTO, error) {
	token, err := tokenService.GetAccessTokenFromContext(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("AuthInterceptor failed to get user from token: %v", err))

		return nil, status.Error(codes.Unauthenticated, "failed to get user from token")
	}
	user, err := tokenService.GetUserFromAcssToken(token)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get user from token: %v", err))
		return nil, status.Error(codes.Unauthenticated, "failed to get user from token")
	}
	return user, nil
}
