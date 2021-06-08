package healthcheck

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("OK"))
}

// StartHealthService starts a health-check service
func StartHealthService() {
	http.HandleFunc("/health", health)
	log.Printf("Staring health check on http://localhost:8090/health")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}
