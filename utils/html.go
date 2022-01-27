package utils

import (
	"bytes"
	"html/template"
	"log"
)

func RenderTemplate(tpl string, data interface{}) []byte {
	var buf bytes.Buffer
	t, err := template.New("").Parse(tpl)
	if err != nil {
		panic(err)
	}

	if err := t.Execute(&buf, data); err == nil {
		return buf.Bytes()
	} else {
		log.Println(err)
		return []byte{}
	}

	return []byte{}
}
