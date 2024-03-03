package victoriametrics

import (
	"log"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

func ListenMetricsServer(bind string) error {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		metrics.WritePrometheus(w, true)
	})

	server := &http.Server{
		Addr:              bind,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
