package updatefw

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	logger "github.com/ecpartan/soap-server-tr069/log"

	"github.com/ecpartan/testing-microservice-soap/config"
)

func parseDirectoriesRegex(htmlContent string) []string {
	// Ищем все ссылки вида: название_папки/ (внутри тега <a>)
	// Паттерн ищет текст между > и </a>, который заканчивается на /
	re := regexp.MustCompile(`<a[^>]*href="([^"]+/)"[^>]*>([^<]+)</a>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var dirs []string
	for _, match := range matches {
		if len(match) >= 3 {
			name := strings.TrimSpace(match[2])
			// Пропускаем родительскую директорию
			if name != "../" && name != ".." && name != "" {
				dirs = append(dirs, name)
			}
		}
	}

	return dirs
}

func getDirectoriesFromURL(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseDirectoriesRegex(string(body)), nil
}

func get_url_update(base_url, fw_name string) string {
	return base_url + fw_name + "/"
}

func updatefw(url string) error {
	filePath := filepath.Join("./scripts/templates/", "upload_fw.json")

	content, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer content.Close()
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	mp := make(map[string]any)

	if json.Unmarshal(data, &mp) != nil {
		return err
	}

	if curr, ok := mp["Script"].(map[string]any); ok {
		if curr1, ok := curr["1"].(map[string]any); ok {
			if uplo, ok := curr1["Upload"].(map[string]any); ok {
				uplo["url"] = url
			}
		}
	}

	return nil
}

func GetUpdatefw() error {

	cfg := config.GetFirmwareConfig()
	if cfg.IsUpdateFw == false {
		return fmt.Errorf("Update FW is disabled")
	}

	if cfg.UpdateFwDir == "" {
		return fmt.Errorf("Update FW dir is empty")
	}
	url := get_url_update(cfg.UpdateFwDir, cfg.Name)

	dirs, err := getDirectoriesFromURL(url)
	if err != nil {
		return err
	}

	sort.Strings(dirs)

	logger.LogDebug("Founded dirs:", len(dirs))

	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]

		if !strings.Contains(d, cfg.Name) && !strings.Contains(d, "_4.0.") {
			continue
		}

		logger.LogDebug("Directory", d)
		d2 := strings.Split(d, "_")[0]
		d3 := strings.Split(d, "_")[1]
		url := url + d + "/" + d2 + "_" + cfg.Name + "_" + d3 + "_" + "nightly.bin"
		fmt.Println("Update url", url)
		return nil
	}

	return fmt.Errorf("Not found dirs for %s", cfg.Name)
}
