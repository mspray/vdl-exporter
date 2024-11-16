package ports

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

// PortCollector collecte les informations sur les ports ouverts et autorisés
type PortCollector struct {
	portsOpenGauge       *prometheus.GaugeVec
	portsAuthorizedGauge *prometheus.GaugeVec
}

// NewPortCollector crée une nouvelle instance de PortCollector
func NewPortCollector() *PortCollector {
	// Métriques pour les ports ouverts
	portsOpenGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ports_open_info",
			Help: "Information about open ports",
		},
		[]string{"protocol", "port", "address", "pid", "process", "interface"},
	)
	prometheus.MustRegister(portsOpenGauge)

	// Métriques pour les ports autorisés (pare-feu)
	portsAuthorizedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ports_authorized_info",
			Help: "Information about authorized ports (e.g., firewalld)",
		},
		[]string{"port", "protocol", "zone"},
	)
	prometheus.MustRegister(portsAuthorizedGauge)

	return &PortCollector{
		portsOpenGauge:       portsOpenGauge,
		portsAuthorizedGauge: portsAuthorizedGauge,
	}
}

// Reset réinitialise les métriques
func (pc *PortCollector) Reset() {
	pc.portsOpenGauge.Reset()
	pc.portsAuthorizedGauge.Reset()
}

// Collect collecte les informations sur les ports ouverts et autorisés
func (pc *PortCollector) Collect() error {
	if err := pc.collectOpenPorts(); err != nil {
		log.WithError(err).Error("Erreur lors de la collecte des ports ouverts")
	}
	if err := pc.collectAuthorizedPorts(); err != nil {
		log.WithError(err).Error("Erreur lors de la collecte des ports autorisés")
	}
	return nil
}

// collectOpenPorts collecte les ports ouverts à l'aide de gopsutil
func (pc *PortCollector) collectOpenPorts() error {
	connections, err := net.Connections("all") // Collecte toutes les connexions TCP et UDP
	if err != nil {
		log.WithError(err).Error("Erreur lors de la récupération des connexions réseau")
		return err
	}

	for _, conn := range connections {
		// Convertir le type de protocole (uint32) en une chaîne lisible
		protocol := getProtocolType(conn.Type)

		// Obtenir le nom du processus à partir du PID
		processName := getProcessName(conn.Pid)

		// Ajouter les métriques
		pc.portsOpenGauge.WithLabelValues(
			protocol,
			fmt.Sprintf("%d", conn.Laddr.Port), // Port local
			conn.Laddr.IP,                      // Adresse locale
			fmt.Sprintf("%d", conn.Pid),        // PID
			processName,                        // Nom du processus
			"",                                 // Interface (optionnelle)
		).Set(1)
	}

	return nil
}

// getProtocolType convertit un uint32 en protocole lisible (tcp ou udp)
func getProtocolType(protocol uint32) string {
	switch protocol {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	default:
		return "unknown"
	}
}

// getProcessName retourne le nom du processus à partir de son PID
func getProcessName(pid int32) string {
	if pid == 0 {
		return ""
	}
	proc, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, err := proc.Name()
	if err != nil {
		return ""
	}
	return name
}

// collectAuthorizedPorts collecte les ports autorisés via firewalld (D-Bus)
func (pc *PortCollector) collectAuthorizedPorts() error {
	zones, err := getFirewallZones()
	if err != nil {
		log.WithError(err).Error("Erreur lors de la récupération des zones firewalld")
		return err
	}

	for _, zone := range zones {
		ports, err := getPortsForZone(zone)
		if err != nil {
			log.WithError(err).Warnf("Erreur lors de la récupération des ports pour la zone %s", zone)
			continue
		}

		for _, portProto := range ports {
			parts := splitPortProtocol(portProto)
			if len(parts) != 2 {
				log.Warnf("Format de port invalide : %s", portProto)
				continue
			}
			port := parts[0]
			protocol := parts[1]
			pc.portsAuthorizedGauge.WithLabelValues(port, protocol, zone).Set(1)
		}
	}

	return nil
}

// getFirewallZones retourne une liste des zones disponibles dans firewalld via D-Bus
func getFirewallZones() ([]string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Error("Erreur de connexion au bus D-Bus")
		return nil, err
	}

	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")
	var zones []string
	err = obj.Call("org.fedoraproject.FirewallD1.getZones", 0).Store(&zones)
	if err != nil {
		log.WithError(err).Error("Erreur lors de la récupération des zones de firewalld")
		return nil, err
	}

	return zones, nil
}

// getPortsForZone retourne les ports autorisés pour une zone spécifique via D-Bus
func getPortsForZone(zone string) ([]string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")
	var ports []string
	err = obj.Call("org.fedoraproject.FirewallD1.zone.listPorts", 0, zone).Store(&ports)
	if err != nil {
		log.WithError(err).Warnf("Erreur lors de la récupération des ports pour la zone %s", zone)
		return nil, err
	}

	return ports, nil
}

// splitPortProtocol sépare une chaîne "port/protocol" en deux parties
func splitPortProtocol(portProto string) []string {
	return strings.Split(portProto, "/")
}
