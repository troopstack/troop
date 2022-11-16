package http

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/troopstack/troop/src/modules/general/utils"
)

func Start() {
	addr := utils.Config().Http.Listen

	if addr == "" {
		return
	}

	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	router := InitRouter()

	s := &http.Server{
		Addr:           addr,
		Handler:        router,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println(" [*] listening http", addr)
	log.Fatalln(s.ListenAndServe())
}
