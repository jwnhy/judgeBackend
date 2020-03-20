package sample

type Type string
type Lang string

const (
	SPJ   = "SPJ"
	Regex = "Regex"
	SQL   = "SQL"
)

const (
	Java     = "Java"
	Postgres = "Postgres"
	SQLite   = "SQLite"
	MySQL    = "MySQL"
	C        = "C"
	CPP      = "CPP"
	Python   = "Python"
)

type Spec struct {
	// Main spec
	Type Type `yaml:"type"`
	Lang Lang `yaml:"lang"`
	// Sub spec
	IsSet      bool   `yaml:"set"`
	IsTrim     bool   `yaml:"trim"`
	Database   string `yaml:"database"`
	DockerFile string `yaml:"dockerfile"`
}
