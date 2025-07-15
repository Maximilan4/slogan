package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Maximilan4/slogan"
)

type handler struct {
	l *slog.Logger
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("my awesome response"))
	h.l.Info("successful write of some text")
}

func main() {
	logger := slog.New(slogan.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: slogan.DefaultReplaceFunc(true, false),
	}))
	// Middleware can be inited without any options
	logMw := slogan.NewRequestLogMiddleware(
		// redeclares logger
		slogan.WithMWLogger(logger),
		// log reqid header value and user agent in start log record
		slogan.WithMWStartLogAttributesFunc(func(req *http.Request) []any {
			return []any{
				slog.String("reqid", req.Header.Get("X-Req-ID")),
				slog.String("ua", req.UserAgent()),
			}
		}),
		// log response status and request URI in end log record
		slogan.WithMWEndLogAttributesFunc(func(req *http.Request, rd *slogan.ResponseData) []any {
			return []any{
				slog.Int64("s", rd.Status),
				slog.String(req.Method, req.RequestURI),
			}
		}),
		// write start log msg
		slogan.WithMWStartMsg("SRT"),
		// disabled default end log msg formatting
		slogan.WithMWEndMsgFunc(nil),
		// rewrite default end msg log
		slogan.WithMWEndMsg("END"),
		// provide cb for start log msg formatting
		slogan.WithMWStartMsgFunc(func(req *http.Request) string {
			return fmt.Sprintf("SRT %s %s", req.Method, req.RequestURI)
		}),
	)

	time.AfterFunc(100*time.Millisecond, func() {
		http.Get("http://127.0.0.1:8000/awesome/resource")
	})

	go func() {
		h := handler{l: logger}

		if err := http.ListenAndServe("127.0.0.1:8000", logMw.Wrap(&h)); err != nil {
			log.Fatal(err)
		}
	}()

	<-time.After(time.Second)
}

// default behaviour:
// level=INFO msg=[START] h=127.0.0.1:8000 met=GET uri=/awesome/resource ua=Go-http-client/1.1 te=<nil> ra=127.0.0.1:64410 cl=0
// level=INFO msg="successful write of some text"
// level=INFO msg="[END=200]" T=98.75µs cl=19

// after opts behavior:
// level=INFO msg="SRT GET /awesome/resource" reqid="" ua=Go-http-client/1.1
// level=INFO msg="successful write of some text"
// level=INFO msg=END s=200 GET=/awesome/resource
