package httpkit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func Serve(ctx context.Context, addr string, chiFunc func(r chi.Router), srvFunc func(srv *http.Server)) {
	// 创建路由
	r := chi.NewRouter()
	// 中间件
	r.Use(Header, Recovery)
	// 自定义
	if chiFunc != nil {
		chiFunc(r)
	}

	// 创建服务
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// 自定义
	if srvFunc != nil {
		srvFunc(srv)
	}

	// 信号监听
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer stop()

	// 启动服务
	go func() {
		fmt.Println("serving on", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("ListenAndServe: %w", err))
		}
	}()

	// 阻塞直到接收到信号
	<-sigCtx.Done()
	fmt.Println("收到终止信号，正在关闭服务...")

	// 关闭服务
	shutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.ErrorContext(ctx, "server shutdown failed", slog.String("error", err.Error()))
	}
}
