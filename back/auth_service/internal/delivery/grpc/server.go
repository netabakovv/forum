package grpc

import (
	"back/auth_service/internal/service"
	"back/auth_service/internal/usecase"
	"back/pkg/errors"
	"back/pkg/logger"
	pb "back/proto"
	"context"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authUC       usecase.AuthUsecaseInterface
	tokenService service.TokenServiceInterface
	logger       logger.Logger
}

func NewAuthServer(authUC usecase.AuthUsecaseInterface, tokenService service.TokenServiceInterface, logger logger.Logger) *AuthServer {
	return &AuthServer{
		authUC:       authUC,
		tokenService: tokenService,
		logger:       logger,
	}
}

// Register создает нового пользователя и возвращает токены доступа
func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		s.logger.Warn("invalid register credentials provided")
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokens, err := s.authUC.Register(ctx, req.Username, req.Password)
	if err != nil {
		s.logger.Error("registration failed",
			logger.NewField("error", err),
			logger.NewField("username", req.Username),
		)
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.RegisterResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(), // используем Unix timestamp
	}, nil
}

// Login аутентифицирует пользователя и возвращает токены доступа
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Проверяем контекст на отмену
	if ctx.Err() != nil {
		s.logger.Error("context error",
			logger.NewField("error", ctx.Err()),
		)
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	s.logger.Info("login request received",
		logger.NewField("username", req.Username),
	)

	if req.Username == "" || req.Password == "" {
		s.logger.Warn("invalid login credentials provided")
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokens, user, err := s.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case errors.ErrInvalidCredentials:
			s.logger.Warn("invalid credentials",
				logger.NewField("username", req.Username),
			)
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.ErrUserNotFound:
			s.logger.Warn("user not found",
				logger.NewField("username", req.Username),
			)
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			s.logger.Error("login failed",
				logger.NewField("error", err),
				logger.NewField("username", req.Username),
			)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	s.logger.Info("login successful",
		logger.NewField("username", req.Username),
		logger.NewField("expires_at", tokens.ExpiresAt),
	)

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
		User: &pb.UserProfileResponse{
			UserId:    user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Unix(),
			IsAdmin:   user.IsAdmin,
		},
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	s.logger.Info("refresh token request received")

	if req.RefreshToken == "" {
		s.logger.Warn("empty refresh token provided")
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokens, err := s.authUC.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("RefreshTokens failed", logger.NewField("error", err))
		switch err {
		case errors.ErrTokenInvalid:
			s.logger.Warn("invalid refresh token")
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		case errors.ErrTokenExpired:
			s.logger.Warn("refresh token expired")
			return nil, status.Error(codes.Unauthenticated, "refresh token expired")
		default:
			s.logger.Error("failed to refresh tokens", logger.NewField("error", err))
			return nil, status.Error(codes.Internal, "failed to refresh tokens")
		}
	}

	s.logger.Info("tokens refreshed successfully", logger.NewField("new_access", tokens.AccessToken[:10]+"…"))

	return &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	s.logger.Info("token validation request received")

	if req.AccessToken == "" {
		s.logger.Warn("empty token provided")
		return &pb.ValidateResponse{
			IsValid: false,
		}, status.Error(codes.InvalidArgument, "token is required")
	}

	claims, err := s.tokenService.ValidateToken(req.AccessToken)
	if err != nil {
		switch err {
		case errors.ErrTokenExpired:
			s.logger.Warn("token expired")
			return &pb.ValidateResponse{
				IsValid: false,
			}, status.Error(codes.Unauthenticated, "token expired")
		case errors.ErrTokenInvalid:
			s.logger.Warn("invalid token")
			return &pb.ValidateResponse{
				IsValid: false,
			}, status.Error(codes.InvalidArgument, "invalid token")
		default:
			s.logger.Error("token validation failed",
				logger.NewField("error", err),
			)
			return &pb.ValidateResponse{
				IsValid: false,
			}, status.Error(codes.Internal, "internal error")
		}
	}

	s.logger.Info("token validated successfully",
		logger.NewField("user_id", claims.UserID),
	)

	return &pb.ValidateResponse{
		UserId:   claims.UserID, // конвертируем int64 в string
		Username: claims.Username,
		IsAdmin:  claims.IsAdmin,
		IsValid:  true,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if ctx == nil {
		return nil, status.Error(codes.InvalidArgument, "контекст не может быть nil")
	}

	if req.AccessToken == "" {
		s.logger.Warn("предоставлен пустой токен")
		return nil, status.Error(codes.InvalidArgument, "требуется токен доступа")
	}

	claims, err := s.tokenService.ValidateToken(req.AccessToken)
	if err != nil {
		switch err {
		case errors.ErrTokenExpired:
			s.logger.Warn("токен просрочен")
			return nil, status.Error(codes.Unauthenticated, "токен просрочен")
		case errors.ErrTokenInvalid:
			s.logger.Warn("недействительный токен")
			return nil, status.Error(codes.InvalidArgument, "недействительный токен")
		default:
			s.logger.Error("ошибка проверки токена",
				logger.NewField("error", err),
			)
			return nil, status.Error(codes.Internal, "внутренняя ошибка")
		}
	}

	s.logger.Info("получен запрос на выход",
		logger.NewField("user_id", claims.UserID),
	)

	if err := s.authUC.RevokeTokens(ctx, claims.UserID); err != nil {
		s.logger.Error("не удалось отозвать токены",
			logger.NewField("error", err),
			logger.NewField("user_id", claims.UserID),
		)
		return nil, status.Error(codes.Internal, "не удалось выполнить выход")
	}

	s.logger.Info("пользователь успешно вышел",
		logger.NewField("user_id", claims.UserID),
	)

	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

func (s *AuthServer) CheckAdminStatus(ctx context.Context, req *pb.CheckAdminRequest) (*pb.CheckAdminResponse, error) {
	if req.UserId == 0 {
		s.logger.Warn("empty user id provided")
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	s.logger.Info("check admin status request received",
		logger.NewField("user_id", req.UserId),
	)

	isAdmin, err := s.authUC.IsAdmin(ctx, req.UserId)
	if err != nil {
		s.logger.Error("failed to check admin status",
			logger.NewField("error", err),
			logger.NewField("user_id", req.UserId),
		)
		return nil, status.Error(codes.Internal, "failed to check admin status")
	}

	return &pb.CheckAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // на время dev можно *
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
