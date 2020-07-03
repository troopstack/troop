package utils

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/chenhg5/collection"
)

type UnCipherText struct {
	AES        string
	GeneralPub string
}

var (
	GeneralNoActiveAddr []string
	GeneralAES          string
	GeneralClientsLock  = new(sync.RWMutex)
	GeneralClients      = map[string]*SingleConnRpcClient{}
	Retry               = 3
)

func initGeneralClient(addr string) *SingleConnRpcClient {
	var client = &SingleConnRpcClient{
		RpcServer: addr,
		Timeout:   time.Duration(Config().General.Timeout) * time.Millisecond,
	}
	GeneralClientsLock.Lock()
	defer GeneralClientsLock.Unlock()
	GeneralClients[addr] = client
	return client
}

func getGeneralClient(addr string) *SingleConnRpcClient {
	GeneralClientsLock.RLock()
	defer GeneralClientsLock.RUnlock()

	if c, ok := GeneralClients[addr]; ok {
		return c
	}
	return nil
}

func CallGeneral(method string, args interface{}, reply interface{}) bool {
	rand.Seed(time.Now().UnixNano())
	var err error
	addresses := Config().General.Addresses
	if len(GeneralNoActiveAddr) >= len(addresses) {
		GeneralNoActiveAddr = []string{}
	}
	for _, i := range rand.Perm(len(addresses)) {
		addr := addresses[i]
		if collection.Collect(GeneralNoActiveAddr).Contains(addr) {
			continue
		}
		c := getGeneralClient(addr)
		if c == nil {
			c = initGeneralClient(addr)
		}
		err = c.Call(method, args, reply)
		if err == nil {
			break
		} else {
			CallOk := false
			RetryCount := Retry
			for RetryCount > 0 {
				err = c.Call(method, args, reply)
				if err == nil {
					CallOk = true
					break
				}
				RetryCount--
			}
			if CallOk {
				break
			} else {
				GeneralNoActiveAddr = append(GeneralNoActiveAddr, addr)
			}
		}
	}
	if err != nil {
		log.Println("call ", method, " fail:", err)
		return false
	}
	return true
}
