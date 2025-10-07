package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	cloudflareIPv4URL = "https://www.cloudflare.com/ips-v4"
	cloudflareIPv6URL = "https://www.cloudflare.com/ips-v6"
	httpTimeout       = 10 * time.Second
)

var cloudflareRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch Cloudflare IPs and update Traefik config",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Cloudflare IP Updater for Traefik"))
		fmt.Println(titleStyle.Render(strings.Repeat("=", 40)))

		newIPs, err := fetchCloudflareIPs()
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error fetching Cloudflare IPs: %v", err)))
			os.Exit(1)
		}

		containersDir := viper.GetString("containers-dir")
		realContainersDir := resolveContainersDir(containersDir)
		traefikConfigPath := filepath.Join(realContainersDir, "traefik", "container-config", "traefik", "traefik.yaml")

		config, err := readTraefikConfig(traefikConfigPath)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error reading traefik config: %v", err)))
			os.Exit(1)
		}

		needsUpdate, updatedConfig := updateCloudflareIPsInConfig(config, newIPs)
		if !needsUpdate {
			fmt.Println(successStyle.Render("✓ Cloudflare IPs are already up to date"))
			fmt.Println(successStyle.Render("\nNo updates needed!"))
			os.Exit(0)
		}

		err = writeTraefikConfig(traefikConfigPath, updatedConfig)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error writing traefik config: %v", err)))
			os.Exit(1)
		}

		fmt.Println(successStyle.Render("\n✓ Cloudflare IPs updated successfully!"))
		restartTraefik()
	},
}

func fetchCloudflareIPs() ([]string, error) {
	fmt.Println(titleStyle.Render("Fetching latest Cloudflare IP ranges..."))
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

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Fetched %d IPv4 and %d IPv6 ranges", len(ipv4Ranges), len(ipv6Ranges))))
	return allRanges, nil
}

func fetchIPs(url string) ([]string, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(string(body)), "\n"), nil
}

func readTraefikConfig(path string) (map[interface{}]interface{}, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("traefik config not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[interface{}]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func updateCloudflareIPsInConfig(config map[interface{}]interface{}, newIPs []string) (bool, map[interface{}]interface{}) {
	cfIPs, ok := config["cloudflare-ips"].(map[interface{}]interface{})
	if !ok {
		fmt.Println(crossMark + " 'cloudflare-ips' section not found in config")
		return false, config
	}

	currentIPsInterface, ok := cfIPs["trustedIPs"].([]interface{})
	if !ok {
		fmt.Println(crossMark + " 'trustedIPs' section not found in 'cloudflare-ips'")
		return false, config
	}

	var currentIPs []string
	for _, ip := range currentIPsInterface {
		if ipStr, ok := ip.(string); ok {
			currentIPs = append(currentIPs, ipStr)
		} else {
			fmt.Println(warningStyle.Render(fmt.Sprintf("Warning: non-string IP found in trustedIPs: %v", ip)))
		}
	}

	sort.Strings(currentIPs)
	sort.Strings(newIPs)

	if strings.Join(currentIPs, ",") == strings.Join(newIPs, ",") {
		return false, config
	}

	var newIPsInterface []interface{}
	for _, ip := range newIPs {
		newIPsInterface = append(newIPsInterface, ip)
	}

	cfIPs["trustedIPs"] = newIPsInterface
	config["cloudflare-ips"] = cfIPs

	return true, config
}

func writeTraefikConfig(path string, config map[interface{}]interface{}) error {
	backupPath := fmt.Sprintf("%s.backup.%s", path, time.Now().Format("20060102_150405"))
	if err := os.Rename(path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}
	fmt.Println(folderMark + " Backup created: " + filePathStyle.Render(backupPath))

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		os.Rename(backupPath, path)
		return fmt.Errorf("failed to write config, backup restored: %v", err)
	}

	fmt.Println(successStyle.Render("✓ Configuration updated successfully"))
	return nil
}

func restartTraefik() {
	fmt.Println(titleStyle.Render("Restarting Traefik container..."))
	cmd := exec.Command("systemctl", "--user", "restart", "traefik.container")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(crossMark + " " + errorStyle.Render(fmt.Sprintf("Failed to restart Traefik: %v\n%s", err, string(output))))
		fmt.Println(infoMark + " Please restart manually: " + "systemctl --user restart traefik.container")
		return
	}
	fmt.Println(successStyle.Render("✓ Traefik restarted successfully"))
}

func init() {
	cloudflareCmd.AddCommand(cloudflareRunCmd)
}
