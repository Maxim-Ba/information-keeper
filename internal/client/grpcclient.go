package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	conn           *grpc.ClientConn
	authClient     proto.AuthClient
	artifactClient proto.ArtifactServiceClient
	TokenManager   *TokenManager
	maxRetries     int
	healthClient   grpc_health_v1.HealthClient
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		conn:           conn,
		authClient:     proto.NewAuthClient(conn),
		artifactClient: proto.NewArtifactServiceClient(conn),
		TokenManager:   &TokenManager{},
		healthClient:   grpc_health_v1.NewHealthClient(conn),
		maxRetries:     2, // TODO from cfg
	}, nil
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
func (c *GRPCClient) HealthCheck(ctx context.Context) error {
	resp, err := c.healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "",
	})

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving, status: %v", resp.Status)
	}

	return nil
}
func (c *GRPCClient) Login(ctx context.Context, login, password string) (*proto.LoginUserResponse, error) {
	req := &proto.LoginUserRequest{
		User: &proto.UserAuthReqDTO{
			Login:    login,
			Password: password,
		},
	}
	logger.Info(fmt.Sprintf("GRPCClient Login login:%s, password:%s", login, password))
	resp, err := c.authClient.Login(ctx, req)

	if err != nil {
		logger.Error(fmt.Sprintf("GRPCClient Login %v", err))
		return resp, err

	}
	logger.Info(fmt.Sprintf("GRPCClient Login resp:%v", resp))
	c.TokenManager.SetTokens(resp.AccessToken, resp.RefreshToken)
	logger.Info(fmt.Sprintf("GRPCClient Login c.TokenManager.GetAccessToken():%s", c.TokenManager.GetAccessToken()))
	go c.SyncLoop(ctx)
	return resp, err
}

func (c *GRPCClient) Register(ctx context.Context, login, password, email string) (*proto.RegistrationUserResponse, error) {
	logger.Info(fmt.Sprintf("GRPCClient Register login:%s, password:%s, email:%s", login, password, email))
	req := &proto.RegistrationUserRequest{
		User: &proto.UserRegistrationReqDTO{
			Login:    login,
			Password: password,
			Email:    email,
		},
	}

	return c.authClient.Register(ctx, req)
}

func (c *GRPCClient) RefreshToken(ctx context.Context) (*proto.LoginUserResponse, error) {
	req := &proto.RefreshTokenRequest{
		RefreshToken: c.TokenManager.GetRefreshToken(),
	}

	return c.authClient.RefreshToken(ctx, req)
}

func (c *GRPCClient) Logout(ctx context.Context) (*proto.LogoutResponse, error) {
	var resp *proto.LogoutResponse
	var err error
	req := &proto.LogoutRequest{
		AccessToken:  c.TokenManager.GetAccessToken(),
		RefreshToken: c.TokenManager.GetRefreshToken(),
	}

	logger.Info("Токены перед logout", "accessToken", req.AccessToken, "refreshToken", req.RefreshToken)

	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.authClient.Logout(ctxWithToken, req)
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}
func (c *GRPCClient) ChangePassword(ctx context.Context) (*proto.ChangePasswordResponse, error) {
	req := &proto.ChangePasswordRequest{
		AccessToken: c.TokenManager.GetAccessToken(),
	}

	return c.authClient.ChangePassword(ctx, req)
}

func (c *GRPCClient) RestorePassword(ctx context.Context, email string) (*proto.RestorePasswordResponse, error) {
	req := &proto.RestorePasswordRequest{
		Email: email,
	}

	return c.authClient.RestorePassword(ctx, req)
}

func (c *GRPCClient) SendEmailConfirmation(ctx context.Context, email string) (*proto.SendEmailConfirmationResponse, error) {
	req := &proto.SendEmailConfirmationRequest{
		Email: email,
	}

	return c.authClient.SendEmailConfirmation(ctx, req)
}

// Artifact methods
func (c *GRPCClient) CreateArtifact(ctx context.Context, req *proto.CreateArtifactRequest) (*proto.CreateArtifactResponse, error) {
	var resp *proto.CreateArtifactResponse
	var err error

	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.artifactClient.CreateArtifact(ctxWithToken, req)
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) GetArtifact(ctx context.Context, artifactID string) (*proto.GetArtifactResponse, error) {
	var resp *proto.GetArtifactResponse
	var err error

	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.artifactClient.GetArtifact(ctxWithToken, &proto.GetArtifactRequest{
			ArtifactId: artifactID,
		})
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) ListArtifacts(ctx context.Context, page, onPage int32, typeFilter *proto.ArtifactTypeEnum) (*proto.ListArtifactsResponse, error) {
	var resp *proto.ListArtifactsResponse
	var err error

	operation := func(ctx context.Context) error {
        ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
        
        // Создаем запрос с опциональным typeFilter
        req := &proto.ListArtifactsRequest{
            Page:   page,
            OnPage: onPage,
        }
        
        // Добавляем typeFilter только если он указан
        if typeFilter != nil {
            req.TypeFilter = typeFilter
        }
        
        resp, err = c.artifactClient.ListArtifacts(ctxWithToken, req)
        return err
    }

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) UpdateArtifact(ctx context.Context, req *proto.UpdateArtifactRequest) (*proto.UpdateArtifactResponse, error) {
	var resp *proto.UpdateArtifactResponse
	var err error

	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.artifactClient.UpdateArtifact(ctxWithToken, req)
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) DeleteArtifact(ctx context.Context, artifactID string) (*proto.DeleteArtifactResponse, error) {
	var resp *proto.DeleteArtifactResponse
	var err error

	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.artifactClient.DeleteArtifact(ctxWithToken, &proto.DeleteArtifactRequest{ArtifactId: artifactID})
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) GetWithOTP(ctx context.Context, artifactID string) (*proto.GetWithOTPResponse, error) {
	var resp *proto.GetWithOTPResponse
	var err error
	req := &proto.GetWithOTPRequest{
		ArtifactId: artifactID,
	}
	operation := func(ctx context.Context) error {
		ctxWithToken := c.withAuthToken(ctx, c.TokenManager.GetAccessToken())
		resp, err = c.artifactClient.GetWithOTP(ctxWithToken, req)
		return err
	}

	if err := c.executeWithTokenRetry(ctx, operation); err != nil {
		return nil, err
	}

	return resp, nil

}

func (c *GRPCClient) Sync(ctx context.Context, clientID string, lastSyncTime int64) (proto.ArtifactService_SyncClient, error) {
	accessToken := c.TokenManager.GetAccessToken()
	ctx = c.withAuthToken(ctx, accessToken)
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}
	// Создаем потоковый клиент
	stream, err := c.artifactClient.Sync(ctx)
	if err != nil {
		return nil, err
	}

	// Отправляем начальный запрос
	req := &proto.SyncRequest{
		ClientId:     clientID,
		LastSyncTime: lastSyncTime,
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	return stream, nil
}
func (c *GRPCClient) StartSync(ctx context.Context, clientID string, eventHandler func(*proto.SyncEvent) error) error {
	var lastSyncTime int64
	// Пытаемся получить время последней синхронизации 
	// lastSyncTime = c.getLastSyncTime(clientID)

	operation := func(ctx context.Context) error {
		stream, err := c.Sync(ctx, clientID, lastSyncTime)
		if err != nil {
			return err
		}

		// Горутина для heartbeat
		go c.sendHeartbeats(ctx, stream, clientID)

		// Обработка входящих событий
		for {
			event, err := stream.Recv()
			if err != nil {
				return err
			}

			// Обновляем время последней синхронизации
			if complete := event.GetSyncComplete(); complete != nil {
				lastSyncTime = complete.SyncTime
				// c.saveLastSyncTime(clientID, lastSyncTime)
			}

			// Передаем событие обработчику
			if err := eventHandler(event); err != nil {
				slog.Error("Error handling sync event", "error", err)
				// Продолжаем получать события несмотря на ошибку обработки
			}
		}
	}
//TODO усли произошел логин то выходим из цыкла

	// Бесконечный цикл переподключения
	for {
		err := c.executeWithTokenRetry(ctx, operation)
		if err != nil && ctx.Err() == nil {
			slog.Error("Sync stream disconnected, reconnecting...", "error", err)
			time.Sleep(5 * time.Second) // Ждем перед переподключением
			continue
		}
		return err
	}
}

func (c *GRPCClient) sendHeartbeats(ctx context.Context, stream proto.ArtifactService_SyncClient, clientID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			heartbeat := &proto.SyncRequest{
				ClientId:     clientID,
				LastSyncTime: time.Now().Unix(),
			}
			if err := stream.Send(heartbeat); err != nil {
				slog.Error("Failed to send heartbeat", "error", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// обертка для работы с синхронизацией
func (c *GRPCClient) SyncWithHandler(
	ctx context.Context,
	clientID string,
	lastSyncTime int64,
	eventHandler func(*proto.SyncEvent) error) error {

	stream, err := c.Sync(ctx, clientID, lastSyncTime)
	if err != nil {
		return err
	}

	// Горутина для отправки heartbeat или дополнительных запросов
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Отправляем heartbeat
				heartbeat := &proto.SyncRequest{
					ClientId:     clientID,
					LastSyncTime: time.Now().Unix(),
				}
				if err := stream.Send(heartbeat); err != nil {
					slog.Error(fmt.Sprintf("Failed to send heartbeat: %v", err))
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Чтение событий из потока
	for {
		event, err := stream.Recv()
		if err != nil {
			return err
		}

		if err := eventHandler(event); err != nil {
			return err
		}
	}
}

// Helper method to add auth token to context
func (c *GRPCClient) withAuthToken(ctx context.Context, token string) context.Context {
	if token == "" {
		slog.Warn("Attempting to create context with empty token")
		return ctx
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})

	// Объединяем с существующими метаданными, если они есть
	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existingMD, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func (c *GRPCClient) refreshTokens(ctx context.Context) error {
	refreshToken := c.TokenManager.GetRefreshToken()
	if refreshToken == "" {
		return errors.New("no refresh token available")
	}

	resp, err := c.authClient.RefreshToken(ctx, &proto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	c.TokenManager.SetTokens(resp.AccessToken, resp.RefreshToken)
	slog.Info("Tokens refreshed successfully")
	return nil
}

// isTokenExpiredError проверяет, является ли ошибка ошибкой истечения токена
func (c *GRPCClient) isTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// Проверяем коды ошибок, которые указывают на истечение токена
	return st.Code() == codes.Unauthenticated ||
		st.Code() == codes.PermissionDenied ||
		(st.Code() == codes.Unknown &&
			(st.Message() == "token expired" ||
				st.Message() == "invalid token" ||
				st.Message() == "access token expired"))
}

// executeWithTokenRetry выполняет операцию с автоматическим обновлением токена при необходимости
func (c *GRPCClient) executeWithTokenRetry(ctx context.Context, operation func(context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		err := operation(ctx)
		if err == nil {
			return nil
		}

		// Если это не ошибка токена или это последняя попытка - возвращаем ошибку
		if !c.isTokenExpiredError(err) || attempt == c.maxRetries-1 {
			return err
		}

		lastErr = err
		slog.Info(fmt.Sprintf("Token expired, attempting refresh (attempt %d)", attempt+1))

		if refreshErr := c.refreshTokens(ctx); refreshErr != nil {
			slog.Error(fmt.Sprintf("Failed to refresh tokens: %v", refreshErr))
			return errors.Join(err, refreshErr)
		}

		//  пауза перед повторной попыткой
		time.Sleep(100 * time.Millisecond)
	}

	return lastErr
}

func (c *GRPCClient) SyncLoop(ctx context.Context) {
	err := c.StartSync(ctx, "client-1", func(event *proto.SyncEvent) error {
		switch e := event.EventType.(type) {
		case *proto.SyncEvent_ArtifactCreated:
			fmt.Printf("New artifact created: %s\n", e.ArtifactCreated.Artifact.Id)
			// Обновить UI
		case *proto.SyncEvent_ArtifactUpdated:
			fmt.Printf("Artifact updated: %s\n", e.ArtifactUpdated.Artifact.Id)
			// Обновить UI
		case *proto.SyncEvent_ArtifactDeleted:
			fmt.Printf("Artifact deleted: %s\n", e.ArtifactDeleted.ArtifactId)
			// Обновить UI
		case *proto.SyncEvent_ClientConnected:
			fmt.Printf("Other client connected: %s\n", e.ClientConnected.ClientId)
		case *proto.SyncEvent_ClientDisconnected:
			fmt.Printf("Other client disconnected: %s\n", e.ClientDisconnected.ClientId)
		}
		return nil
	})
	if err != nil {
		slog.Error("Sync failed", "error", err)
	}
}
