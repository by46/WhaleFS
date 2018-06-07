package common

type Config struct {
	Host     string `default:":8080"`
	Master   string
	Debug    bool   `default:"false"`
	LogLevel string `default:"error"` // fatal, error, warning, info, debug
	LogHome  string `default:"log"`
	Meta     string `default:"http://scpodb01:8091/"`
	Bucket   string `default:"dfis"`
}
