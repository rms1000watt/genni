package generator

import (
	log "github.com/sirupsen/logrus"
)

func Generator(cfg Config) {
	log.Debug("Starting Generator")
	defer log.Debug("Done Generator")

	structs, err := Parse(cfg)
	if err != nil {
		log.Error("Failed parsing: ", err)
		return
	}

	if err := Write(structs, cfg); err != nil {
		log.Error("Failed writing: ", err)
		return
	}
}
