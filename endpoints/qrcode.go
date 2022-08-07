package endpoints

import (
	"github.com/skip2/go-qrcode"
	"log"
	"net/http"
	"strings"
)

func QrCodeEndpoint() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		address := strings.TrimPrefix(request.RequestURI, "/address/")
		png, err := qrcode.Encode(address, qrcode.Medium, 256)
		if err != nil {
			log.Println(err)
			return
		}

		writer.WriteHeader(200)
		writer.Header().Add("Cache-Control", "max-age=86400")
		writer.Header().Add("Content-Type", "image/png")
		_, err = writer.Write(png)
		if err != nil {
			return
		}
	}
}
