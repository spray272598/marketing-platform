package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
)

// TestGRPCServerWrapperLifecycle 验证 gRPC server 适配器能真正接入 kratos.App
// 的生命周期：Start 监听端口、Endpoint 返回 grpc:// 形式的地址、Stop 优雅退出。
// 这是把 *grpc.Server 适配为 transport.Server 的核心契约。
func TestGRPCServerWrapperLifecycle(t *testing.T) {
	gs := grpc.NewServer()
	w := &grpcServerWrapper{Server: gs, addr: "127.0.0.1:0"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { _ = w.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)

	ep, err := w.Endpoint()
	if err != nil {
		t.Fatalf("Endpoint error: %v", err)
	}
	if ep.Scheme != "grpc" {
		t.Fatalf("expected grpc scheme, got %q", ep.Scheme)
	}

	if err := w.Stop(ctx); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}
