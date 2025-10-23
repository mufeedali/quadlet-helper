package backup

// BackupType represents the type of backup tool
type BackupType string

const (
	BackupTypeRsync  BackupType = "rsync"
	BackupTypeRestic BackupType = "restic"
	BackupTypeRclone BackupType = "rclone"
)

// Config represents a backup configuration
type Config struct {
	Name          string        `yaml:"name"`
	Type          BackupType    `yaml:"type"`
	Schedule      string        `yaml:"schedule"`
	Source        []string      `yaml:"source"`
	Destination   Destination   `yaml:"destination"`
	Options       Options       `yaml:"options,omitempty"`
	Verification  Verification  `yaml:"verification,omitempty"`
	Retention     Retention     `yaml:"retention,omitempty"`
	Notifications Notifications `yaml:"notifications,omitempty"`
	Hooks         Hooks         `yaml:"hooks,omitempty"`
	Environment   []string      `yaml:"environment,omitempty"`
}

// Destination varies by backup type
type Destination struct {
	Remote     string `yaml:"remote,omitempty"`     // For rclone
	Path       string `yaml:"path,omitempty"`       // For rsync
	Repository string `yaml:"repository,omitempty"` // For restic
}

// Options contains backup-specific options
type Options struct {
	// Rclone options
	Transfers      int      `yaml:"transfers,omitempty"`
	Checkers       int      `yaml:"checkers,omitempty"`
	BandwidthLimit string   `yaml:"bandwidth_limit,omitempty"`
	Exclude        []string `yaml:"exclude,omitempty"`

	// Rsync options
	Archive  bool `yaml:"archive,omitempty"`
	Compress bool `yaml:"compress,omitempty"`
	Delete   bool `yaml:"delete,omitempty"`

	// Restic options
	PasswordFile string `yaml:"password_file,omitempty"`
	KeepDaily    int    `yaml:"keep_daily,omitempty"`
	KeepWeekly   int    `yaml:"keep_weekly,omitempty"`
}

// Verification settings
type Verification struct {
	Enabled    bool   `yaml:"enabled"`
	AutoVerify bool   `yaml:"auto_verify"`
	Method     string `yaml:"method,omitempty"` // check, checksum, size, cryptcheck
}

// Retention settings
type Retention struct {
	KeepDays    int `yaml:"keep_days,omitempty"`
	KeepDaily   int `yaml:"keep_daily,omitempty"`
	KeepWeekly  int `yaml:"keep_weekly,omitempty"`
	KeepMonthly int `yaml:"keep_monthly,omitempty"`
}

// Notifications settings
type Notifications struct {
	Enabled   bool        `yaml:"enabled"`
	OnFailure bool        `yaml:"on_failure"`
	OnSuccess bool        `yaml:"on_success"`
	Email     EmailConfig `yaml:"email,omitempty"`
}

// EmailConfig for email notifications
type EmailConfig struct {
	To   string `yaml:"to,omitempty"`
	From string `yaml.omitempty:"from,omitempty"`
}

// Hooks for pre/post backup scripts
type Hooks struct {
	PreBackup  string `yaml:"pre_backup,omitempty"`
	PostBackup string `yaml:"post_backup,omitempty"`
	OnFailure  string `yaml:"on_failure,omitempty"`
}
