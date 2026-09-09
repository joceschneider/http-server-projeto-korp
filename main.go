package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Response struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

var requestsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "projeto_korp_requests_total",
		Help: "Total de requisicoes recebidas pelo endpoint /projeto-korp.",
	},
)

func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	requestsTotal.Inc()

	if r.Method != http.MethodGet {
		http.Error(w, "Metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	response := Response{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erro ao gerar resposta", http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "UP",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erro ao gerar resposta", http.StatusInternalServerError)
	}
}

func main() {
	prometheus.MustRegister(requestsTotal)

	http.HandleFunc("/projeto-korp", projetoKorpHandler)
	http.HandleFunc("/health", healthHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Servidor iniciado na porta 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
