package internalgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cronnoss/banners-rotation/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger *logger.Logger
}

func NewLoggingInterceptor(logg *logger.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logg,
	}
}

func (l *LoggingInterceptor) logRequest(ctx context.Context, method string, req interface{}) {
	_ = ctx
	l.logger.Info(fmt.Sprintf("Received request: method=%s, request=%v", method, req))
}

func (l *LoggingInterceptor) logResponse(ctx context.Context, method string, resp interface{}) {
	_ = ctx
	l.logger.Info("Sent %s response: %+v", method, resp)
}

func (l *LoggingInterceptor) UnaryServerInterceptor(
	ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	l.logRequest(ctx, info.FullMethod, req)
	startTime := time.Now()
	resp, err := handler(ctx, req)
	latency := time.Since(startTime)

	statusCode := 0
	if err != nil {
		statusCode = int(status.Code(err))
		l.logger.Error("Error: %v", err)
	} else {
		l.logResponse(ctx, info.FullMethod, resp)
	}

	l.logger.Info("INFO [%s] { ClientIPAddress:%s user-agent:%s StartAt:%s gRPCMethod:%s StatusCode:%d Latency:%s}",
		time.Now().Format("2006-01-02 15:04:05"),
		getIP(ctx),
		getUserAgent(ctx),
		startTime.Format("2006-01-02 15:04:05.999999 -0700 MST"),
		methodFromFullMethod(info.FullMethod),
		statusCode,
		latency,
	)
	return resp, err
}

func getIP(ctx context.Context) string {
	var clientIP string
	peer, ok := peer.FromContext(ctx)
	if ok {
		clientIP = peer.Addr.String()
	} else {
		clientIP = ""
	}
	return clientIP
}

func methodFromFullMethod(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

func getUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	userAgent := "unknown"
	if ok {
		userAgent = md["user-agent"][0]
	}
	return userAgent
}
