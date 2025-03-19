package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func main() {
	logWriter := io.Writer(os.Stderr)
	if os.Getenv("ENV") == "local" {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	}
	log := zerolog.New(logWriter).With().Timestamp().Str("service", "gtdlog").Logger()

	c := alice.New()
	// inject the logger in request context
	c = c.Append(hlog.NewHandler(log))
	// log info after request is completed
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("request completed")
	}))
	// add ip, user agent and request id to the logger
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RequestIDHandler("req_id", "X-Request-Id"))

	http.Handle("/", c.Then(http.HandlerFunc(handleRoot)))
	http.Handle("/hello", c.Then(http.HandlerFunc(handleHello)))

	serverAddress := ":9172"
	log.Info().Str("address", serverAddress).Msg("starting server")
	log.Fatal().Err(http.ListenAndServe(serverAddress, nil)).Msg("startup failed")
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)

	logger.Info().Str("endpoint", "hello").Msg("hello endpoint accessed")
	response := generateHelloResponse(r.Context())
	logger.Info().Interface("response", response).Msg("sending hello response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("let me guess, you just came to say hello?"))

}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	logger := hlog.FromRequest(r)
	logger.Info().Msg("root endpoint accessed")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("welcome to the API"))
}

type HelloResponse struct {
	Message string `json:"message"`
}

func generateHelloResponse(ctx context.Context) HelloResponse {
	zerolog.Ctx(ctx).Info().Msg("i just came to say hello")

	return HelloResponse{Message: "hello, world!"}
}
