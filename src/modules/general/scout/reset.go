package scout

import (
	"log"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/rmq"
)

func SendHandshakeMessage() {
	ScoutMessage := model.ScoutMessage{
		Type: "handshake",
	}
	_, err := rmq.AmqpServer.PutIntoQueue("scout", "scout", ScoutMessage, 0)

	if err != nil {
		log.Printf("send handshake message error: %s", err)
	}
}
