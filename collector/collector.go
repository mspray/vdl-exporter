package collector

import (
	"context"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"vdl-exporter/config"
	"vdl-exporter/packages"
	"vdl-exporter/ports"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector est responsable de la collecte des métriques
type Collector struct {
	cfg              *config.Config
	mu               sync.Mutex
	packageCollector *packages.PackageCollector
	portCollector    *ports.PortCollector
}

// NewCollector crée une nouvelle instance de Collector
func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		cfg:              cfg,
		packageCollector: packages.NewPackageCollector(cfg),
		portCollector:    ports.NewPortCollector(),
	}
}

// Collect exécute une collecte unique des métriques
func (c *Collector) Collect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Info("Démarrage de la collecte des données...")

	// Réinitialiser les métriques
	c.packageCollector.Reset()
	c.portCollector.Reset()

	// Collecter les données
	if err := c.packageCollector.Collect(); err != nil {
		log.WithError(err).Error("Erreur lors de la collecte des paquets installés")
		return err
	}
	if err := c.portCollector.Collect(); err != nil {
		log.WithError(err).Error("Erreur lors de la collecte des ports")
		return err
	}

	log.Info("Collecte des données terminée.")
	return nil
}

// Start démarre la collecte périodique des données
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.ScrapeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.Collect(); err != nil {
					log.WithError(err).Error("Erreur lors de la collecte périodique")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// MetricsHandler retourne le handler HTTP pour les métriques Prometheus
func (c *Collector) MetricsHandler() http.Handler {
	return promhttp.Handler()
}
