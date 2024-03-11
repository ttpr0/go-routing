package main

import (
	"net/http"
	"os"

	"golang.org/x/exp/slog"
)

var MANAGER *RoutingManager

func main() {
	logger := slog.New(NewLogHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Initializing GoRouting Server...")

	config := ReadConfig("./config.yml")
	MANAGER = NewRoutingManager("./graphs", config)

	app := http.DefaultServeMux

	MapPost(app, "/v0/routing", HandleRoutingRequest)
	MapPost(app, "/v0/routing/draw/create", HandleCreateContextRequest)
	MapPost(app, "/v0/routing/draw/step", HandleRoutingStepRequest)
	MapPost(app, "/v0/isoraster", HandleIsoRasterRequest)
	MapPost(app, "/v1/matrix", HandleMatrixRequest)

	http.ListenAndServe(":5002", nil)
}
