package main

import (
	"context"
	"io"
	"strings"
	"sync"

	"golang.org/x/exp/slog"
)

type LogHandler struct {
	h   slog.Handler
	mu  *sync.Mutex
	out io.Writer
}

func NewLogHandler(o io.Writer, opts *slog.HandlerOptions) *LogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &LogHandler{
		out: o,
		h: slog.NewTextHandler(o, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: nil,
		}),
		mu: &sync.Mutex{},
	}
}

func (h *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogHandler{h: h.h.WithAttrs(attrs), out: h.out, mu: h.mu}
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	return &LogHandler{h: h.h.WithGroup(name), out: h.out, mu: h.mu}
}

func (h *LogHandler) Handle(ctx context.Context, r slog.Record) error {

	formattedTime := r.Time.Format("2006/01/02 15:04:05")

	//add time and message to values
	strs := []string{formattedTime, r.Level.String(), r.Message, "\n"}

	if r.NumAttrs() != 0 {
		r.Attrs(func(a slog.Attr) bool {
			strs = append(strs, a.Value.String())
			return true
		})
	}

	result := strings.Join(strs, " ")
	b := []byte(result)

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.out.Write(b)

	return err

}
