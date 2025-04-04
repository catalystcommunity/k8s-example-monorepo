package cmd

import (
	"fmt"
	"net/http"

	"github.com/catalystcommunity/app-utils-go/errorutils"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/handlers"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store/postgres_store"
	"github.com/gammazero/workerpool"
)

var Server *http.ServeMux

func Serve() error {
	// set stores
	store.AppStore = postgres_store.PostgresStore

	// init stores and defer any functions we need to
	deferredStoreFuncs := initStores()
	for _, deferredFunc := range deferredStoreFuncs {
		defer deferredFunc()
	}
	
	// Create the handler with routes
	handler := handlers.NewRouter()
	
	// Log startup information
	logging.Log.Infof("Starting HTTP server on port %d", config.Port)
	
	// Start the HTTP server
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), handler)
	
	// ListenAndServe always eventually errors out, so we log it and return it
	errorutils.LogOnErr(nil, "ListenAndServe exited with: ", err)
	return err
}

func initStores() []func() {
	// initialize stores using a worker pool to speed up startup
	pool := workerpool.New(5)
	deferredFunctions := []func(){}

	pool.Submit(func() {
		deferredFunc, err := store.AppStore.Initialize()
		errorutils.PanicOnErr(nil, "error initializing app store", err)
		if deferredFunc != nil {
			deferredFunctions = append(deferredFunctions, deferredFunc)
		}
		logging.Log.Info("app store initialized")
	})

	pool.StopWait()
	return deferredFunctions
}
