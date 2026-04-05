package cloudflare

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

const (
	cloudflareIPv4URL = "https://www.cloudflare.com/ips-v4"
	cloudflareIPv6URL = "https://www.cloudflare.com/ips-v6"
	httpTimeout       = 10 * time.Second
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch Cloudflare IPs and update Traefik config",
	RunE: func(c *cobra.Command, args []string) error {
		fmt.Println(shared.TitleStyle.Render("Cloudflare IP Updater for Traefik"))
		fmt.Println(shared.TitleStyle.Render(strings.Repeat("=", 40)))

		newIPs, err := fetchCloudflareIPs()
		if err != nil {
			return cmdutil.Wrap(err, "fetching Cloudflare IPs")
		}

		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)
		traefikConfigPath := shared.TraefikConfigPath(realContainersPath)

		config, err := readTraefikConfig(traefikConfigPath)
		if err != nil {
			return cmdutil.Wrap(err, "reading traefik config")
		}

		needsUpdate, updatedConfig, err := updateCloudflareIPsInConfig(config, newIPs)
		if err != nil {
			return cmdutil.Wrap(err, "updating cloudflare IPs in traefik config")
		}
		if !needsUpdate {
			fmt.Println(shared.SuccessStyle.Render("✓ Cloudflare IPs are already up to date"))
			fmt.Println(shared.SuccessStyle.Render("\nNo updates needed!"))
			return nil
		}

		err = writeTraefikConfig(traefikConfigPath, updatedConfig)
		if err != nil {
			return cmdutil.Wrap(err, "writing traefik config")
		}

		fmt.Println(shared.SuccessStyle.Render("\n✓ Cloudflare IPs updated successfully!"))
		return restartTraefik()
	},
}

func fetchCloudflareIPs() ([]string, error) {
	fmt.Println(shared.TitleStyle.Render("Fetching latest Cloudflare IP ranges..."))
	var allRanges []string

	ipv4Ranges, err := fetchIPs(cloudflareIPv4URL)
	if err != nil {
		return nil, err
	}
	allRanges = append(allRanges, ipv4Ranges...)

	ipv6Ranges, err := fetchIPs(cloudflareIPv6URL)
	if err != nil {
		return nil, err
	}
	allRanges = append(allRanges, ipv6Ranges...)

	fmt.Println(shared.SuccessStyle.Render(fmt.Sprintf("✓ Fetched %d IPv4 and %d IPv6 ranges", len(ipv4Ranges), len(ipv6Ranges))))
	return allRanges, nil
}

func fetchIPs(url string) ([]string, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from %s: %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(body)), "\n"), nil
}

func readTraefikConfig(path string) (map[string]any, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("traefik config not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]any
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config:\n%w", err)
	}

	return config, nil
}

func updateCloudflareIPsInConfig(config map[string]any, newIPs []string) (bool, map[string]any, error) {
	cfSection, ok := valueAsStringMap(config["cloudflare-ips"])
	if !ok {
		return false, config, fmt.Errorf("'cloudflare-ips' section not found in config")
	}

	currentIPs, ok := valueAsStringSlice(cfSection["trustedIPs"])
	if !ok {
		return false, config, fmt.Errorf("'cloudflare-ips.trustedIPs' must be a list of strings")
	}

	sortedCurrent := append([]string(nil), currentIPs...)
	sortedNew := append([]string(nil), newIPs...)
	sort.Strings(sortedCurrent)
	sort.Strings(sortedNew)

	if strings.Join(sortedCurrent, ",") == strings.Join(sortedNew, ",") {
		return false, config, nil
	}

	cfSection["trustedIPs"] = append([]string(nil), newIPs...)
	config["cloudflare-ips"] = cfSection

	return true, config, nil
}

func writeTraefikConfig(path string, config map[string]any) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config: %w", err)
	}

	backupPath := fmt.Sprintf("%s.backup.%s", path, time.Now().Format("20060102_150405"))
	if err := os.Rename(path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}
	fmt.Println(shared.FolderMark + " Backup created: " + shared.FilePathStyle.Render(backupPath))

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		_ = enc.Close()
		_ = os.Rename(backupPath, path)
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	_ = enc.Close()

	if err := os.WriteFile(path, buf.Bytes(), info.Mode().Perm()); err != nil {
		_ = os.Rename(backupPath, path)
		return fmt.Errorf("failed to write config, backup restored: %v", err)
	}

	fmt.Println(shared.SuccessStyle.Render("✓ Configuration updated successfully"))
	return nil
}

func valueAsStringMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		converted := make(map[string]any, len(typed))
		for key, nested := range typed {
			keyString, ok := key.(string)
			if !ok {
				return nil, false
			}
			converted[keyString] = nested
		}
		return converted, true
	default:
		return nil, false
	}
}

func valueAsStringSlice(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...), true
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			itemString, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, itemString)
		}
		return result, true
	default:
		return nil, false
	}
}

func restartTraefik() error {
	fmt.Println(shared.TitleStyle.Render("Restarting Traefik container..."))
	if _, err := systemd.Restart("traefik.container"); err != nil {
		fmt.Println(shared.CrossMark + " " + shared.ErrorStyle.Render(fmt.Sprintf("Failed to restart Traefik: %v", err)))
		fmt.Println(shared.InfoMark + " Please restart manually: " + "systemctl --user restart traefik.container")
		return cmdutil.Wrap(err, "restarting Traefik")
	}
	fmt.Println(shared.SuccessStyle.Render("✓ Traefik restarted successfully"))
	return nil
}
