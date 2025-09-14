package server

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/Maxim-Ba/information-keeper/internal/server/dto"
	"github.com/Maxim-Ba/information-keeper/internal/server/services"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	pb "github.com/Maxim-Ba/information-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authService AuthServiceInterface

	pb.UnimplementedAuthServer
}

type ArtifactServer struct {
	pb.UnimplementedArtifactServiceServer
}
type HealthServer struct {
	pb.UnimplementedHealthServer
}
type PasswordServiceInterface interface {
	ChangePassword(ctx context.Context, data *dto.ChangePassword) error
}

type AuthServiceInterface interface {
	Login(ctx context.Context, login, password string) (*services.JWTToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*services.JWTToken, error)
	Logout(ctx context.Context, token *services.JWTToken) error
	Register(ctx context.Context, login, email, password string) (*services.JWTToken, error)
	SendEmailConfirmation(ctx context.Context, email string) error
	PasswordServiceInterface
}

type GRPCServer struct {
	server          *grpc.Server
	artifactService interface{}
}

func NewGRPCServer(authService AuthServiceInterface, artifactService interface{}) *GRPCServer {
	grpcServer := grpc.NewServer()

	authServer := &AuthServer{authService: authService}
	artifactServer := &ArtifactServer{}
	healthServer := health.NewServer()

	pb.RegisterAuthServer(grpcServer, authServer)
	pb.RegisterArtifactServiceServer(grpcServer, artifactServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	healthServer.SetServingStatus("auth", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("artifact", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING) 
	return &GRPCServer{
		server: grpcServer,

		artifactService: artifactService,
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
	// TODO: Реализовать создание артефакта
	slog.Info("AuthServer CreateArtifact")

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
	// TODO: Реализовать получение списка артефактов
	logger.Info("AuthServer ListArtifacts")

	return &pb.ListArtifactsResponse{
		Artifacts: []*pb.Artifact{},
	}, nil
}

func (s *ArtifactServer) UpdateArtifact(ctx context.Context, req *pb.UpdateArtifactRequest) (*pb.UpdateArtifactResponse, error) {
	// TODO: Реализовать обновление артефакта
	slog.Info("AuthServer UpdateArtifact")

	return &pb.UpdateArtifactResponse{}, nil
}

func (s *ArtifactServer) DeleteArtifact(ctx context.Context, req *pb.DeleteArtifactRequest) (*pb.DeleteArtifactResponse, error) {
	// TODO: Реализовать удаление артефакта
	slog.Info("AuthServer DeleteArtifact")

	return &pb.DeleteArtifactResponse{}, nil
}

func (s *ArtifactServer) GetWithOTP(ctx context.Context, req *pb.GetWithOTPRequest) (*pb.GetWithOTPResponse, error) {
	// TODO: Реализовать получение с OTP
	slog.Info("AuthServer GetWithOTP")

	return &pb.GetWithOTPResponse{}, nil
}

func (s *ArtifactServer) Sync(stream pb.ArtifactService_SyncServer) error {
	// TODO: Реализовать синхронизацию в реальном времени
	slog.Info("AuthServer Sync")

	for {
		_, err := stream.Recv()
		if err != nil {
			return err
		}

		// Отправляем событие синхронизации
		event := &pb.SyncEvent{}
		if err := stream.Send(event); err != nil {
			return err
		}
	}
}
