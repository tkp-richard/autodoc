package autodoc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// ginResponseRecorder writes to both a ResponseRecorder and the original ResponseWriter
type ginResponseRecorder struct {
	gin.ResponseWriter
	recorder     *httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *ginResponseRecorder) Header() http.Header {
	return r.recorder.Header()
}

func (r *ginResponseRecorder) Write(b []byte) (int, error) {
	fmt.Printf(">> debug >> string(b): %#v\n", string(b))
	r.recorder.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *ginResponseRecorder) WriteString(s string) (n int, err error) {
	fmt.Printf(">> debug >> s: %#v\n", s)
	r.recorder.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func (r *ginResponseRecorder) WriteHeader(statusCode int) {
	// TODO: temp fix for sse
	if statusCode == -1 {
		statusCode = 200
	}
	r.recorder.WriteHeader(statusCode)
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ginResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func createGinResponseRecorder(w gin.ResponseWriter) *ginResponseRecorder {
	return &ginResponseRecorder{
		ResponseWriter: w,
		recorder:       httptest.NewRecorder(),
		closeChannel:   make(chan bool, 1),
	}
}

func createTestGinContext(c *gin.Context) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := createGinResponseRecorder(c.Writer)
	c.Writer = rec
	return c, rec.recorder
}

func (r *Recorder) RecordGin(h gin.HandlerFunc, opts ...RecordOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		c, rec := createTestGinContext(c)

		if c.Request.URL.Path == "" {
			p := r.Path
			re := regexp.MustCompile(`{(.*?)}`)
			matches := re.FindAllString(r.Path, -1)
			for _, m := range matches {
				p = strings.ReplaceAll(p, m, c.Param(strings.Trim(m, "{}")))
			}
			c.Request.URL.Path = p
		}

		h(c)

		r.record(c.Request, rec.Result(), opts...)
	}
}
