package expvar

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.uber.org/zap"
)

// Expvar provide our basic publishing.
type Expvar struct {
	log    *zap.Logger
	server http.Server
	data   map[string]interface{}
	mu     sync.Mutex
}

// New starts a service for consuming the raw expvar stats.
func New(log *zap.Logger, host string, route string, readTimeout, writeTimeout time.Duration) *Expvar {
	mux := httptreemux.New()
	exp := Expvar{
		log: log,
		server: http.Server{
			Addr:           host,
			Handler:        mux,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}

	mux.Handle("GET", route, exp.handler)

	go func() {
		log.Info("expvar: API Listening", zap.String("host", host))
		if err := exp.server.ListenAndServe(); err != nil {
			log.Error("ERROR", zap.Error(err))
		}
	}()

	return &exp
}

// Stop shuts down the service.
func (exp *Expvar) Stop(shutdownTimeout time.Duration) {
	exp.log.Info("expvar: Start shutdown...")
	defer exp.log.Info("expvar: Completed")

	// Create context for Shutdown call.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Asking listener to shutdown and load shed.
	if err := exp.server.Shutdown(ctx); err != nil {
		exp.log.Error("ERROR: Graceful shutdown did not complete", zap.Error(err), zap.Duration("shutdownTimeout", shutdownTimeout))
		if err := exp.server.Close(); err != nil {
			exp.log.Error("ERROR: Could not stop http server", zap.Error(err))
		}
	}
}

// Publish is called by the publisher goroutine and saves the raw stats.
func (exp *Expvar) Publish(data map[string]interface{}) {
	exp.mu.Lock()
	{
		exp.data = data
	}
	exp.mu.Unlock()
}

// handler is what consumers call to get the raw stats.
func (exp *Expvar) handler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var data map[string]interface{}
	exp.mu.Lock()
	{
		data = exp.data
	}
	exp.mu.Unlock()

	if err := json.NewEncoder(w).Encode(data); err != nil {
		exp.log.Error("ERROR", zap.Error(err))
	}

	log.Printf("expvar : (%d) : %s %s -> %s",
		http.StatusOK,
		r.Method, r.URL.Path,
		r.RemoteAddr,
	)
}
