package server

import (
	"context"
	"log/slog"
	"net"

	pb "github.com/Maxim-Ba/information-keeper/pkg/proto"
	"google.golang.org/grpc"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
}

type ArtifactServer struct {
	pb.UnimplementedArtifactServiceServer
}

type GRPCServer struct {
	server *grpc.Server
	authService  interface{}
	artifactService interface{}
}

func NewGRPCServer(authService, artifactService interface{}) *GRPCServer {
	grpcServer := grpc.NewServer()

	authServer := &AuthServer{}
	artifactServer := &ArtifactServer{}

	pb.RegisterAuthServer(grpcServer, authServer)
	pb.RegisterArtifactServiceServer(grpcServer, artifactServer)

	return &GRPCServer{
		server: grpcServer,
		authService: authService,
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

// Реализация методов Auth сервиса
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	// TODO: Реализовать логику аутентификации
	slog.Info("AuthServer Login")
	return &pb.LoginUserResponse{
		RefreshToken: "mock_refresh_token",
		AccessToken:  "mock_access_token",
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginUserResponse, error) {
	// TODO: Реализовать обновление токена
	slog.Info("AuthServer RefreshToken")
	return &pb.LoginUserResponse{
		RefreshToken: "new_refresh_token",
		AccessToken:  "new_access_token",
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// TODO: Реализовать выход
	slog.Info("AuthServer Logout")

	return &pb.LogoutResponse{
		
	}, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegistrationUserRequest) (*pb.RegistrationUserResponse, error) {
	// TODO: Реализовать регистрацию
	slog.Info("AuthServer Register")

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
	slog.Info("AuthServer ListArtifacts")

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
