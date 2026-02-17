package observe

import (
	"net/http"

	prom_middleware "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitPromMetrics() *prom_middleware.ServerMetrics {
	srvMetrics := prom_middleware.NewServerMetrics(
		prom_middleware.WithServerHandlingTimeHistogram(),
	)
	prom_client.MustRegister(srvMetrics)
	return srvMetrics
}

func InitPrometheusHTTPServer(port string) *http.Server {
	mux := http.DefaultServeMux
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return srv
}
