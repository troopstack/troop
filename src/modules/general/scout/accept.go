package scout

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/cache"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/utils"
)

type UnCipherText struct {
	AES        string
	GeneralPub string
}

func AcceptScout(args *model.ScoutInfo) ([]byte, error) {
	// 读General公钥
	GeneralPub, err := ioutil.ReadFile(utils.GeneralPubFilename)
	if err != nil {
		log.Fatalf("General Public read failed：%s", err)
		return nil, errors.New("General Public read failed")
	}
	UnCipherText := UnCipherText{
		AES:        utils.AES,
		GeneralPub: string(GeneralPub),
	}

	data, err := json.Marshal(UnCipherText)

	if err != nil {
		log.Fatalf("Json marshaling failed：%s", err)
		return nil, errors.New("Json marshaling failed")
	}
	// 加密数据
	EncryptData, err := utils.RsaEncrypt(data, []byte(args.PubKey))
	if err != nil {
		return nil, err
	}

	// MySQL存储
	database.UpdateScout(args)

	// 清理缓存的Scout
	cache.Scouts.Delete(args.Hostname)

	return EncryptData, nil
}
