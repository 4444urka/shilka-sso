package auth

import (
	"context"
	"errors"
	ssov1 "github.com/4444urka/shilka-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"shilka-sso/internal/services/auth"
)

// Auth методы, которые необходимо реализовать хэндлерам
type Auth interface {
	Login(
		ctx context.Context,
		username string,
		password string,
		appID int,
	) (token string, err error)

	Register(
		ctx context.Context,
		username string,
		password string,
	) (userID int64, err error)

	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

// RegisterServer Регестрирует сервер с методами, описанными в Auth interface
func RegisterServer(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *ServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {

	//  Валидация
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil

}

func (s *ServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {

	// Валидация
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword())

	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *ServerAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {

	// Валидация
	if err := validateIsAdmin(req); err != nil {
		if errors.Is(err, auth.ErrInvalidUserId) {
			return nil, status.Error(codes.InvalidArgument, "invalid user id")
		}

		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())

	if err != nil {
		//TODO: обработать ошибку
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

// Функции для валидации

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetUsername() == "" {
		return status.Errorf(codes.InvalidArgument, "username is required")
	}

	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Errorf(codes.InvalidArgument, "appId is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetUsername() == "" {
		return status.Errorf(codes.InvalidArgument, "username is empty")
	}

	if req.GetPassword() == "" {
		return status.Errorf(codes.InvalidArgument, "password is empty")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Errorf(codes.InvalidArgument, "userId is empty")
	}

	return nil
}
