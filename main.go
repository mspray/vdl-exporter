package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"vdl-exporter/collector"
	"vdl-exporter/config"
)

func main() {
	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erreur lors du chargement de la configuration : %v", err)
	}

	// Créer un collecteur avec la configuration
	coll := collector.NewCollector(cfg)

	// Contexte pour gérer l'arrêt du programme
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Démarrer la collecte périodique des données
	coll.Start(ctx)

	// Démarrer le serveur HTTP pour exposer les métriques et gérer les requêtes
	go startHTTPServer(cfg, coll)

	// Gestion des signaux pour un arrêt propre
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Info("Arrêt du programme...")
	cancel()
}

// startHTTPServer démarre le serveur HTTP pour exposer les métriques et gérer les requêtes de rafraîchissement
func startHTTPServer(cfg *config.Config, coll *collector.Collector) {
	// Endpoint pour rafraîchir la collecte manuellement
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if err := coll.Collect(); err != nil {
			http.Error(w, "Erreur lors de la collecte des données", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Collecte des données déclenchée")
	})

	// Endpoint pour les métriques Prometheus
	http.Handle("/metrics", coll.MetricsHandler())

	log.Infof("Serveur en écoute sur %s", cfg.ListenAddress)
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		log.WithError(err).Fatal("Erreur lors du démarrage du serveur HTTP")
	}
}
