package rpc

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"

	"github.com/troopstack/troop/src/modules/general/utils"
)

type Scout int

func Start() {
	addr := utils.Config().Rpc.Listen

	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	s := rpc.NewServer()
	err := s.Register(new(Scout))
	if err != nil {
		log.Fatalln("RPC Register error:", err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("listen error:", err)
	} else {
		log.Println(" [*] listening", addr)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("listener accept fail:", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}
		go s.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
