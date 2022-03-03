package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// OpenTracing :
func (mw *Middleware) OpenTracing(typ string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			r := c.Request()

			requestBody := ""

			operation := "HTTP " + r.Method + " " + r.URL.Path

			span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), operation)
			defer span.Finish()

			span.SetTag("type", typ)

			ext.HTTPMethod.Set(span, string(r.Method))
			ext.HTTPUrl.Set(span, r.URL.String())
			ext.Component.Set(span, "xenotification")
			c.Set("OpenTracingSpan", ctx)

			requestHeader, err := json.Marshal(c.Request().Header)
			if err == nil {
				span.LogKV("http.request.header", string(requestHeader))
			}

			isContentTypeJSON := strings.Contains(c.Request().Header.Get("Content-Type"), "application/json") || c.Request().Header.Get("Content-Type") == ""
			if c.Request().Body != nil && isContentTypeJSON {
				reqBody := sortJSON(c)
				requestBody = reqBody.String()
				c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody.Bytes()))
			}
			span.LogKV("http.request.body", requestBody)

			// Response
			resBody := new(bytes.Buffer)
			if isContentTypeJSON {
				mw := io.MultiWriter(c.Response().Writer, resBody)
				writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
				c.Response().Writer = writer
			}

			defer func() {
				responseHeader, err := json.Marshal(c.Response().Header())
				if err == nil {
					span.LogKV("http.response.header", string(responseHeader))
				}

				if c.Response().Size < 50000 {
					responseBody := resBody.String()
					span.LogKV("http.response.body", responseBody)
				}

				span = setTag(span, c)
				statusCode := c.Response().Status

				if statusCode > 399 {
					span.SetTag("error", true)
				} else {
					span.SetTag("error", false)
				}
				ext.HTTPStatusCode.Set(span, uint16(c.Response().Status))
			}()

			return next(c)
		}
	}
}

func setTag(sp opentracing.Span, c echo.Context) opentracing.Span {
	// header := c.Request().Header

	remoteIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	requestID := c.Response().Header().Get("X-Request-Id")

	sp.SetTag("request_id", requestID)
	sp.SetTag("remote_ip", remoteIP)
	sp.SetTag("user_agent", userAgent)

	return sp
}

func sortJSON(c echo.Context) *bytes.Buffer {
	body := new(bytes.Buffer)

	if c.Request().Body != nil { // Read
		reqBodyByte, _ := ioutil.ReadAll(c.Request().Body)
		body.Write(reqBodyByte)
	}

	return body
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
