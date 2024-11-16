package packages

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	"vdl-exporter/config"

	"github.com/prometheus/client_golang/prometheus"
)

// PackageCollector collecte les informations sur les paquets installés
type PackageCollector struct {
	cfg                   *config.Config
	packageInstalledGauge *prometheus.GaugeVec
}

// NewPackageCollector crée une nouvelle instance de PackageCollector
func NewPackageCollector(cfg *config.Config) *PackageCollector {
	packageInstalledGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "package_installed_info",
			Help: "Information about installed packages",
		},
		[]string{"name", "version"},
	)
	prometheus.MustRegister(packageInstalledGauge)

	return &PackageCollector{
		cfg:                   cfg,
		packageInstalledGauge: packageInstalledGauge,
	}
}

// Reset réinitialise les métriques
func (pc *PackageCollector) Reset() {
	pc.packageInstalledGauge.Reset()
}

// Collect collecte les informations sur les paquets installés
func (pc *PackageCollector) Collect() error {
	if _, err := exec.LookPath("rpm"); err == nil {
		return pc.collectRPMPackages()
	} else if _, err := exec.LookPath("dpkg-query"); err == nil {
		return pc.collectDPKGPackages()
	}
	log.Warn("Ni RPM ni DPKG n'est disponible sur ce système")
	return nil
}

func (pc *PackageCollector) collectRPMPackages() error {
	cmd := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}|%{VERSION}\n")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		version := parts[1]
		if _, excluded := pc.cfg.ExcludePackages[name]; excluded {
			continue // Exclure le paquet
		}
		pc.packageInstalledGauge.WithLabelValues(name, version).Set(1)
	}
	return nil
}

func (pc *PackageCollector) collectDPKGPackages() error {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}|${Version}\n")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		version := parts[1]
		if _, excluded := pc.cfg.ExcludePackages[name]; excluded {
			continue // Exclure le paquet
		}
		pc.packageInstalledGauge.WithLabelValues(name, version).Set(1)
	}
	return nil
}
