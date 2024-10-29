package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	authgrpc "shilka-sso/internal/grpc/auth"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.RegisterServer(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun Запускает сервер и роняет приложенияе если сервер не запускается
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run Запускает gRPC сервер
func (a *App) Run() error {
	const operation = "grpcApp.Run"

	log := a.log.With(slog.String("operation", operation),
		slog.Int("port", a.port))

	log.Info("starting gRPC server")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))

	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	log.Info("gRPC server started", slog.String("address", lis.Addr().String()))

	if err := a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

// Stop сервер выполняет оставшиеся запросы а затем останавливается
func (a *App) Stop() error {
	const operation = "grpcApp.Stop"

	a.log.With(slog.String("operation", operation)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()

	a.log.Info("gRPC server stopped", slog.Int("port", a.port))

	return nil
}
