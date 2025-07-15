# slogan
Provides set of extension funcs for default log/slog package.

## Install
```shell
go get github.com/Maximilan4/slogan
```

## Usage
Package provide two basic wrappers of default slog handlers:
- slogan.TextHandler
- slogan.JSONHandler
They are provide two main features:
- additional logging values from context.Context
- increase/decrease log level for specific context

### Handlers
```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Maximilan4/slogan"
)

// ReqIdCKey - example type for storing value in ctx
type ReqIdCKey struct{}

// handler will call GetKey and use string value as an atribute key
func (r ReqIdCKey) GetKey() string {
	return "reqid"
}

func main() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		// DefaultReplaceFunc - can disable time and rewrite source key value
		ReplaceAttr: slogan.DefaultReplaceFunc(true, true),
	}

	// third argument - context keys, must implement ContextKeyGetter
	handler := slogan.NewTextHandler(os.Stdout, &opts, ReqIdCKey{})
	logger := slog.New(handler)

	logger.Info("first log")
	// create a new context with value and log it
	ctx := context.WithValue(context.Background(), ReqIdCKey{}, "example")
	logger.InfoContext(ctx, "log with context value")

	// debug skipping via global handler settings
	logger.DebugContext(ctx, "you will not see this")

	// decrease log level for a specific call, wrap ctx above with debug value
	debugCtx := slogan.DebugContext(ctx)
	logger.DebugContext(debugCtx, "now you see me")

	// increase log level for a specific call
	warnCtx := slogan.WarnContext(ctx)
	logger.WarnContext(warnCtx, "filter info and debug calls")
	logger.InfoContext(warnCtx, "skip info")
	logger.DebugContext(warnCtx, "skip debug")
}
```
```shell
#output
level=INFO source=main.main:31 msg="first log"
level=INFO source=main.main:34 msg="log with context value" reqid=example
level=DEBUG source=main.main:41 msg="now you see me" reqid=example
level=WARN source=main.main:45 msg="filter info and debug calls" reqid=example
```


### Attributes
Package provides some attributes and values funcs:
- slogan.Pointer - prints a value after dereference or prints nil
- slogan.Map - make a group from map. Nested values are supported
- slogan.Slice - make a group from slice. 
- slogan.Any - extended slog.Any by functions above
```go 
package main

import (
	"log/slog"
	"os"

	"github.com/Maximilan4/slogan"
)

func main() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		// DefaultReplaceFunc - can disable time and rewrite source key value
		ReplaceAttr: slogan.DefaultReplaceFunc(true, true),
	}

	// third argument - context keys, must implement ContextKeyGetter
	handler := slogan.NewTextHandler(os.Stdout, &opts)
	logger := slog.New(handler)

	exampleValue := "example string value"
	logger.Info("log a pointer",
		slog.Any("slog", &exampleValue), // produce a pointer value like slog=0x1400010af30
		slogan.Pointer("slogan", &exampleValue)) // will dereference p value
    // level=INFO source=main.main:23 msg="log a pointer" slog=0x1400010af30 slogan="example string value"
	
	mValue := map[string]any{
		"strkey": "this is a string",
		"intkey": 123,
		"nested": map[int]int{
			1: 2,
			3: 4,
		},
	}

	logger.Info("log a map value",
		slog.Any("slog", mValue),     // default short format output map[intkey:123 nested:map[1:2 3:4] strkey:this is a string]
		slogan.Map("slogan", mValue), // will produce group with name sloganmap, nested maps are supported
	)
	// level=INFO source=main.main:36 msg="log a map value" slog="map[intkey:123 nested:map[1:2 3:4] strkey:this is a string]" slogan.strkey="this is a string" slogan.intkey=123 slogan.nested.1=2 slogan.nested.3=4
	
	parts := []string{"my", "split", "string"}
	logger.Info("log a slice value",
		slog.Any("slog", parts),                         // default output slog="[my split string]"
		slogan.Slice("slogan", parts, false),            // will produce group with indexes keys
		slogan.Slice("unwrap", []string{"value"}, true), // will unwrap slice single value to a key like unwrap=value
	)
	// level=INFO source=main.main:42 msg="log a slice value" slog="[my split string]" slogan.0=my slogan.1=split slogan.2=string unwrap=value
}
```

### Middleware for default handler
Package also provide a start/end log middleware for basic http.Handler:
```go
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
		handler := func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			writer.Write([]byte("my awesome response"))
		}

		if err := http.ListenAndServe("127.0.0.1:8000", logMw.Wrap(http.HandlerFunc(handler))); err != nil {
			log.Fatal(err)
		}
	}()

	<-time.After(time.Second)
}
```
```shell
#output
level=INFO msg="SRT GET /awesome/resource" reqid="" ua=Go-http-client/1.1
level=INFO msg=END s=200 GET=/awesome/resource
```
Visit examples for code run. [examples](examples)
