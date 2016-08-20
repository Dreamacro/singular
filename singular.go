package singular

import (
	log "github.com/Sirupsen/logrus"
)

// CheckError print error log
func CheckError(msg string, err error) {
	if err != nil {
		log.Errorf("%s: %v", msg, err)
	}
}

// PassOrFatal print error and fatal
func PassOrFatal(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}
