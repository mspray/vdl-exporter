package utils

import (
	"net"
	"path/filepath"
	"strings"
)

// ParseAddressPort extrait l'adresse et le port à partir d'une chaîne donnée.
// Prend en charge les formats IPv4 et IPv6.
func ParseAddressPort(addrPort string) (string, string) {
	if strings.Contains(addrPort, "[") {
		// Format IPv6
		parts := strings.Split(addrPort, "]:")
		if len(parts) == 2 {
			address := strings.TrimPrefix(parts[0], "[")
			port := parts[1]
			return address, port
		}
	} else {
		// Format IPv4
		parts := strings.Split(addrPort, ":")
		if len(parts) == 2 {
			address := parts[0]
			port := parts[1]
			return address, port
		}
	}
	return "", "" // Retourne des chaînes vides en cas d'erreur de parsing
}

// ParsePIDProcess extrait le PID et le nom du processus à partir d'une chaîne donnée.
// Gère les formats comme users:(pid=123,exe=/path/to/exe).
func ParsePIDProcess(pidProc string) (string, string) {
	if pidProc == "-" || pidProc == "" {
		return "", "" // Aucun processus associé
	}

	pid := ""
	process := ""

	if strings.HasPrefix(pidProc, "users:(") {
		pidProc = strings.TrimPrefix(pidProc, "users:(")
		pidProc = strings.TrimSuffix(pidProc, ")")
	}

	parts := strings.Split(pidProc, ",")
	for _, part := range parts {
		if strings.Contains(part, "pid=") {
			pid = strings.TrimPrefix(part, "pid=")
			pid = strings.Trim(pid, "\"")
		} else if strings.Contains(part, "exe=") {
			process = strings.TrimPrefix(part, "exe=")
			process = strings.Trim(process, "\"")
			process = filepath.Base(process) // Extrait uniquement le nom de l'exécutable
		}
	}

	return pid, process
}

// GetInterface détermine le nom de l'interface réseau associée à une adresse IP donnée.
// Parcourt toutes les interfaces disponibles sur le système.
func GetInterface(address string) string {
	if address == "" {
		return "" // Adresse vide
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "" // Retourne une chaîne vide en cas d'erreur
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // Ignore les erreurs sur cette interface
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.String() == address {
				return iface.Name // Retourne le nom de l'interface correspondante
			}
		}
	}
	return "" // Aucun match trouvé
}
