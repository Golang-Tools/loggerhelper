package loggerhelper

import (
	"context"
	"log/slog"

	"github.com/Golang-Tools/optparams"
)

// WithHandlerMiddleware 追加任意handler中间件,用于替代logrus版本的hook
func WithHandlerMiddleware(mws ...HandlerMiddleware) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.HandlerMiddlewares = append(append([]HandlerMiddleware{}, o.HandlerMiddlewares...), mws...)
	})
}

// WithHandlerMiddlewares 整体设置handler中间件列表(覆盖已有),传空参数可清空
func WithHandlerMiddlewares(mws ...HandlerMiddleware) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.HandlerMiddlewares = append([]HandlerMiddleware{}, mws...)
	})
}

// WithFanoutOutput 内置中间件:把每条日志同时分发给主输出与额外的handler(多路输出)
// extra可为不同格式/不同目标的handler
func WithFanoutOutput(extra ...slog.Handler) optparams.Option[Options] {
	return WithHandlerMiddleware(func(next slog.Handler) slog.Handler {
		return &fanoutHandler{handlers: append([]slog.Handler{next}, extra...)}
	})
}

// WithLevelRoute 内置中间件:达到threshold及以上的日志转交给route处理,其余仍由主输出处理
func WithLevelRoute(threshold slog.Level, route slog.Handler) optparams.Option[Options] {
	return WithHandlerMiddleware(func(next slog.Handler) slog.Handler {
		return &routeHandler{threshold: threshold, def: next, route: route}
	})
}

// fanoutHandler 将记录分发给多个handler
type fanoutHandler struct {
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, hd := range h.handlers {
		if hd != nil && hd.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hd := range h.handlers {
		if hd == nil || !hd.Enabled(ctx, r.Level) {
			continue
		}
		if err := hd.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, hd := range h.handlers {
		if hd != nil {
			hs[i] = hd.WithAttrs(attrs)
		}
	}
	return &fanoutHandler{handlers: hs}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(h.handlers))
	for i, hd := range h.handlers {
		if hd != nil {
			hs[i] = hd.WithGroup(name)
		}
	}
	return &fanoutHandler{handlers: hs}
}

// routeHandler 按级别阈值将记录路由到不同handler
type routeHandler struct {
	threshold slog.Level
	def       slog.Handler
	route     slog.Handler
}

func (h *routeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.threshold {
		if h.route != nil && h.route.Enabled(ctx, level) {
			return true
		}
	}
	return h.def != nil && h.def.Enabled(ctx, level)
}

func (h *routeHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.threshold && h.route != nil && h.route.Enabled(ctx, r.Level) {
		return h.route.Handle(ctx, r)
	}
	if h.def != nil && h.def.Enabled(ctx, r.Level) {
		return h.def.Handle(ctx, r)
	}
	return nil
}

func (h *routeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := &routeHandler{threshold: h.threshold}
	if h.def != nil {
		nh.def = h.def.WithAttrs(attrs)
	}
	if h.route != nil {
		nh.route = h.route.WithAttrs(attrs)
	}
	return nh
}

func (h *routeHandler) WithGroup(name string) slog.Handler {
	nh := &routeHandler{threshold: h.threshold}
	if h.def != nil {
		nh.def = h.def.WithGroup(name)
	}
	if h.route != nil {
		nh.route = h.route.WithGroup(name)
	}
	return nh
}
