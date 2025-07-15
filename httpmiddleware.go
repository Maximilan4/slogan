package slogan

import (
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Maximilan4/gentypes"
)

type (
	RequestLogMiddleware struct {
		l            *slog.Logger
		startMsg     string
		endMsg       string
		startAttrs   func(req *http.Request) []any
		endAttrs     func(req *http.Request, rd *ResponseData) []any
		startMsgFunc func(req *http.Request) string
		endMsgFunc   func(req *http.Request, rd *ResponseData) string
	}

	ResponseData struct {
		Status, Written int64
		StartedAt       time.Time
		Writer          http.ResponseWriter
	}

	RLMOption gentypes.GenericOption[RequestLogMiddleware]
)

// WithMWLogger - option for setting logger to middleware
func WithMWLogger(l *slog.Logger) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.l = l
	}
}

// WithMWStartLogAttributesFunc - custom func option for collecting START log attributes from http.Request
func WithMWStartLogAttributesFunc(f func(req *http.Request) []any) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.startAttrs = f
	}
}

// WithMWEndLogAttributesFunc custom func option for collecting END log attributes from http.Request and ResponseData
func WithMWEndLogAttributesFunc(f func(req *http.Request, rd *ResponseData) []any) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.endAttrs = f
	}
}

// WithMWEndMsg - option redeclare END log msg
func WithMWEndMsg(msg string) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.endMsg = msg
	}
}

// WithMWStartMsg - option redeclare START log msg
func WithMWStartMsg(msg string) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.startMsg = msg
	}
}

// WithMWEndMsgFunc - provides cb for formatting END msg (func has priority over given EndMsg)
func WithMWEndMsgFunc(f func(req *http.Request, rd *ResponseData) string) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.endMsgFunc = f
	}
}

// WithMWStartMsgFunc - provides cb for formatting START msg (func has priority over given StartMsg)
func WithMWStartMsgFunc(f func(req *http.Request) string) RLMOption {
	return func(r *RequestLogMiddleware) {
		r.startMsgFunc = f
	}
}

// NewRequestLogMiddleware -
func NewRequestLogMiddleware(opts ...RLMOption) *RequestLogMiddleware {
	rlm := RequestLogMiddleware{
		l:          slog.Default(),
		startAttrs: CollectDefaultStartLogAttrs,
		startMsg:   "[START]",
		endMsg:     "[END]",
		endMsgFunc: defaultEndMessageFunc,
	}

	for opt := range slices.Values(opts) {
		opt(&rlm)
	}

	return &rlm
}

// defaultEndMessageFunc - prepare END log msg in [END=<status>] format
func defaultEndMessageFunc(_ *http.Request, rd *ResponseData) string {
	var builder strings.Builder
	builder.WriteString("[END=")
	builder.WriteString(strconv.FormatInt(rd.Status, 10))
	builder.WriteByte(']')

	return builder.String()
}

// CollectDefaultStartLogAttrs - collect default START log attributes by http.Request
func CollectDefaultStartLogAttrs(req *http.Request) []any {
	return []any{
		slog.String("h", req.Host),
		slog.String("met", req.Method),
		slog.String("uri", req.RequestURI),
		slog.String("ua", req.UserAgent()),
		Slice("te", req.TransferEncoding, true),
		slog.String("ra", req.RemoteAddr),
		slog.Int64("cl", req.ContentLength),
	}
}

// CollectDefaultEndLogAttrs - collect default END log attributes by ResponseData
func CollectDefaultEndLogAttrs(_ *http.Request, rd *ResponseData) []any {
	return []any{
		slog.Duration("T", time.Since(rd.StartedAt)),
		slog.Int64("cl", rd.Written),
	}
}

func (rlm *RequestLogMiddleware) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request == nil {
			return
		}
		startTime := time.Now()
		ctx := request.Context()

		var startMsg string
		if rlm.startMsgFunc != nil {
			startMsg = rlm.startMsgFunc(request)
		} else {
			startMsg = rlm.startMsg
		}

		rlm.l.InfoContext(request.Context(), startMsg, rlm.collectStartAttrs(request)...)
		h.ServeHTTP(writer, request)

		if writer == nil {
			rlm.l.InfoContext(ctx, rlm.endMsg)
			return
		}

		// reflect is only way to get status and written bytes from private http.response
		v := reflect.Indirect(reflect.ValueOf(writer))
		rd := ResponseData{
			Status:    v.FieldByName("status").Int(),
			Written:   v.FieldByName("written").Int(),
			Writer:    writer,
			StartedAt: startTime,
		}

		var endMsg string
		if rlm.endMsgFunc != nil {
			endMsg = rlm.endMsgFunc(request, &rd)
		} else {
			endMsg = rlm.endMsg
		}

		endAttrs := rlm.collectEndAttr(request, &rd)
		if endAttrs == nil {
			rlm.l.InfoContext(ctx, endMsg)
			return
		}

		rlm.l.InfoContext(ctx, endMsg, endAttrs...)
	})
}

func (rlm *RequestLogMiddleware) collectStartAttrs(req *http.Request) []any {
	if rlm.startAttrs != nil {
		return rlm.startAttrs(req)
	}

	return CollectDefaultStartLogAttrs(req)
}

func (rlm *RequestLogMiddleware) collectEndAttr(req *http.Request, rd *ResponseData) []any {
	if rlm.endAttrs != nil {
		return rlm.endAttrs(req, rd)
	}

	return CollectDefaultEndLogAttrs(req, rd)
}
