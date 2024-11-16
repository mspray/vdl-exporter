package config

import (
    "flag"
    "io/ioutil"
    "strings"
    "time"
)

// Config contient la configuration du programme
type Config struct {
    ScrapeInterval     time.Duration
    ExcludePackages    map[string]struct{}
    ListenAddress      string
    ExcludePackagesFile string
}

// LoadConfig charge la configuration à partir des paramètres en ligne de commande et des fichiers
func LoadConfig() (*Config, error) {
    cfg := &Config{}

    flag.DurationVar(&cfg.ScrapeInterval, "scrape.interval", 5*time.Minute, "Intervalle entre chaque collecte de métriques (ex: 15s, 12h)")
    flag.StringVar(&cfg.ExcludePackagesFile, "exclude.packages-file", "", "Chemin vers le fichier contenant les paquets à exclure")
    flag.StringVar(&cfg.ListenAddress, "web.listen-address", ":9100", "Adresse d'écoute pour l'interface web et la télémétrie")
    flag.Parse()

    // Charger les paquets à exclure
    excludePackages, err := loadExcludePackages(cfg.ExcludePackagesFile)
    if err != nil {
        return nil, err
    }
    cfg.ExcludePackages = excludePackages

    return cfg, nil
}

// loadExcludePackages charge la liste des paquets à exclure depuis un fichier
func loadExcludePackages(filePath string) (map[string]struct{}, error) {
    packagesMap := make(map[string]struct{})
    if filePath == "" {
        return packagesMap, nil
    }

    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        pkg := strings.TrimSpace(line)
        if pkg != "" && !strings.HasPrefix(pkg, "#") {
            packagesMap[pkg] = struct{}{}
        }
    }
    return packagesMap, nil
}
