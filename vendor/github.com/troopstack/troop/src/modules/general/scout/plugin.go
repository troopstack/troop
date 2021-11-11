package scout

import (
	"encoding/json"
	"log"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/rmq"
	"github.com/troopstack/troop/src/modules/general/utils"
)

func SendUpdatePluginMessage() {
	Task := &model.ScoutPluginRequest{
		Action:  "update_plugins",
		Plugins: utils.Plugins,
	}
	data, err := json.Marshal(Task)

	ScoutMessage := model.ScoutMessage{
		Type: "plugin",
		Data: []byte(utils.AES_CBC_Encrypt(data, utils.AES)),
	}
	_, err = rmq.AmqpServer.PutIntoQueue("scout", "scout", ScoutMessage, 0)

	if err != nil {
		log.Printf("send update plugin message error: %s", err)
	}
}
