package generator

import (
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"
)

func Write(genni Genni, cfg Config) (err error) {
	tpl, err := template.New("data").Funcs(funcMap).Parse(dataTpl)
	if err != nil {
		log.Debug("Failed parsing data template: ", err)
		return
	}
	if err = tpl.Execute(os.Stdout, genni); err != nil {
		log.Debug("Failed executing template: ", err)
		return
	}
	return
}
