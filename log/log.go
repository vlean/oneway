package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	//log.SetLevel(log.InfoLevel)
}
