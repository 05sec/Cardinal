package db

import "github.com/jinzhu/gorm"

// If the version is different from the database record, it will ask user for cleaning the database.
const VERSION = "20201127"

// Manager is a gorm model for database table `managers`.
type Manager struct {
	gorm.Model

	Name     string
	Password string `json:"-"`
	IsCheck  bool
	Token    string // For single sign-on
}

// Team is a gorm model for database table `teams`.
type Team struct {
	gorm.Model

	Name      string
	Password  string `json:"-"`
	Logo      string
	Score     float64
	SecretKey string
}

// Token is a gorm model for database table `tokens`.
// It used to store team token.
type Token struct {
	gorm.Model

	TeamID uint
	Token  string
}

// Challenge is a gorm model for database table `challenges`, used to store the challenges like Web1, Pwn1.
type Challenge struct {
	gorm.Model
	Title           string
	BaseScore       int
	AutoRefreshFlag bool
	Command         string
}

// DownAction is a gorm model for database table `down_actions`.
type DownAction struct {
	gorm.Model

	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       int
}

// AttackAction is a gorm model for database table `attack_actions`.
// Used to store the flag submitted record.
type AttackAction struct {
	gorm.Model

	TeamID         uint // Victim's team ID
	GameBoxID      uint // Victim's gamebox ID
	ChallengeID    uint // Victim's challenge ID
	AttackerTeamID uint // Attacker's Team ID
	Round          int
}

// Flag is a gorm model for database table `flags`.
// All the flags will be generated before the competition start and save in this table.
type Flag struct {
	gorm.Model

	TeamID      uint
	GameBoxID   uint
	ChallengeID uint
	Round       int
	Flag        string
}

// GameBox is a gorm model for database table `gameboxes`.
type GameBox struct {
	gorm.Model
	ChallengeID uint
	TeamID      uint

	IP          string
	Port        string
	SSHPort     string
	SSHUser     string
	SSHPassword string
	Description string
	Visible     bool
	Score       float64 // The score can be negative.
	IsDown      bool
	IsAttacked  bool
}

// Score is a gorm model for database table `scores`.
// Every action (checkdown, attacked...) will be created a score record, and the total score will be calculated by SUM(`score`).
type Score struct {
	gorm.Model

	TeamID    uint
	GameBoxID uint
	Round     int
	Reason    string
	Score     float64 `gorm:"index"`
}

// Bulletin is a gorm model for database table `bulletins`.
type Bulletin struct {
	gorm.Model

	Title   string
	Content string
}

// BulletinRead gorm model, used to store the bulletin is read by a team.
type BulletinRead struct {
	gorm.Model

	TeamID     uint
	BulletinID uint
}

// Log is a gorm model for database table `logs`.
type Log struct {
	gorm.Model

	Level   int // 0 - Normal, 1 - Warning, 2 - Important
	Kind    string
	Content string
}

// WebHook is a gorm model for database table `webhook`, used to store the webhook.
type WebHook struct {
	gorm.Model

	URL   string
	Type  string
	Token string

	Retry   int
	Timeout int
}

// DynamicConfig is the config which is stored in database.
// So it's a GORM model for users can edit it anytime.
type DynamicConfig struct {
	gorm.Model `json:"-"`

	Key   string
	Value string
	Kind  int8
}
