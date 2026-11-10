package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Dokhoyan/common/pkg/closer"
	"github.com/Dokhoyan/marketplace/internal/config"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	loggerMaxSize    = 10
	loggerMaxBackups = 3
	loggerMaxAge     = 3
)

var configPath string
var logLevel = flag.String("level", "info", "log level for logger")

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

type App struct {
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		closer.CloseAll()
		closer.Wait()
		return nil, err
	}
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- a.runHTTPServer()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			_ = a.httpServer.Close()
			return fmt.Errorf("failed to shut down HTTP server: %w", err)
		}
		if err := <-serverErrChan; err != nil {
			return err
		}
		logger.Info("server stopped")
		return nil
	case err := <-serverErrChan:
		if err != nil {
			logger.Error("server error occurred, shutting down", zap.Error(err))
		}
		return err
	}
}

func (a *App) initDeps(ctx context.Context) error {
	flag.Parse()

	inits := []func(context.Context) error{
		a.initLogger,
		a.initConfig,
		a.initServiceProvider,
		a.initHTTPServer,
	}
	for _, f := range inits {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initConfig(_ context.Context) error {
	if err := config.Load(configPath); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func (a *App) initServiceProvider(ctx context.Context) error {
	serviceProvider, err := newServiceProvider(ctx)
	if err != nil {
		return fmt.Errorf("failed to init service provider: %w", err)
	}
	a.serviceProvider = serviceProvider
	return nil
}

func (a *App) initHTTPServer(_ context.Context) error {
	mux := http.NewServeMux()
	auth := a.serviceProvider.AuthMiddleware()
	a.serviceProvider.AuthImpl().RegisterRoutes(mux, auth)
	a.serviceProvider.ProductImpl().RegisterRoutes(mux, auth)
	a.serviceProvider.UserImpl().RegisterRoutes(mux, auth)

	var handler http.Handler = mux
	handler = middleware.RequestLogger(handler)

	httpConfig := a.serviceProvider.HTTPConfig()
	a.httpServer = &http.Server{
		Addr:              httpConfig.Address(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	level, err := a.getLevel()
	if err != nil {
		return err
	}
	if err := os.MkdirAll("logs", 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	logger.Init(a.getCore(level))
	return nil
}

func (a *App) runHTTPServer() error {
	logger.Info("starting server", zap.String("addr", a.serviceProvider.HTTPConfig().Address()))
	err := a.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}
	return nil
}

func (a *App) getCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)
	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    loggerMaxSize,
		MaxBackups: loggerMaxBackups,
		MaxAge:     loggerMaxAge,
	}
	closer.Add(logFile.Close)
	file := zapcore.AddSync(logFile)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)
}

func (a *App) getLevel() (zap.AtomicLevel, error) {
	var level zapcore.Level
	if err := level.Set(*logLevel); err != nil {
		return zap.AtomicLevel{}, fmt.Errorf("failed to set log level: %w", err)
	}
	return zap.NewAtomicLevelAt(level), nil
}
