package backup

import (
	"strings"
	"testing"
)

func TestGetServiceTemplateSanitizesAndQuotesValues(t *testing.T) {
	config := &Config{
		Environment: []string{"FOO=value with spaces", "MULTILINE=one\ntwo"},
		Verification: Verification{
			Enabled:    true,
			AutoVerify: true,
		},
		Retention: Retention{KeepDays: 1},
	}

	template := GetServiceTemplate("/usr/local/bin/qh", "bad\nname", config)

	checks := []string{
		"Description=Backup: bad name",
		`ExecStart="/usr/local/bin/qh" backup run "bad\nname"`,
		`Environment="FOO=value with spaces"`,
		`Environment="MULTILINE=one\ntwo"`,
		`ExecStartPost="/usr/local/bin/qh" backup verify "bad\nname"`,
		`ExecStopPost="/usr/local/bin/qh" backup cleanup "bad\nname"`,
	}

	for _, check := range checks {
		if !strings.Contains(template, check) {
			t.Fatalf("GetServiceTemplate() missing %q in:\n%s", check, template)
		}
	}
}

func TestGetTimerTemplateSanitizesName(t *testing.T) {
	template, err := GetTimerTemplate("bad\nname", "daily")
	if err != nil {
		t.Fatalf("GetTimerTemplate() error = %v", err)
	}

	if !strings.Contains(template, "Description=Backup timer for bad name") {
		t.Fatalf("GetTimerTemplate() did not sanitize description:\n%s", template)
	}
	if !strings.Contains(template, "Unit=bad name-backup.service") {
		t.Fatalf("GetTimerTemplate() did not sanitize unit name:\n%s", template)
	}
}
