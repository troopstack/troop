package rpc

import (
	"errors"
	"log"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/cache"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/scout"
	"github.com/troopstack/troop/src/modules/general/utils"
)

func (t *Scout) GeneralInitData(args *model.NullRpcRequest, reply *model.RpcInitDataResponse) error {
	reply.Plugins = utils.Plugins
	reply.IgnoreCommands = utils.Config().IgnoreCommand.Commands
	return nil
}

func (t *Scout) Handshake(args *model.ScoutHandRequest, reply *model.ScoutHandResponse) error {
	if args.Hostname == "" || args.PubKey == "" {
		return errors.New("parameter error")
	}

	// 检查DB是否已存在此HostName
	_, dbExists := database.IsExistScout(args.Hostname)

	if !dbExists {
		// 检查被拒绝的Scout缓存列表是否包含此Hostname
		data, exists := cache.Scouts.Get(args.Hostname)
		if exists && data.Status == "denied" {
			return errors.New("rejected")
		}
	}

	status := "unaccepted"
	aesCi := utils.AES_CBC_Encrypt([]byte(args.AES), utils.AESKey)

	var ScoutInfo = &model.ScoutInfo{
		Hostname:     args.Hostname,
		IP:           args.IP,
		ScoutVersion: args.ScoutVersion,
		Type:         args.Type,
		PubKey:       args.PubKey,
		Status:       status,
		AES:          aesCi,
		OS:           args.OS,
		Tags:         args.Tags,
		Plugins:      args.Plugins,
	}
	log.Println("Handshake: ", (*model.ScoutInfo).String(ScoutInfo))

	if !dbExists {
		// 未信任该Scout
		err := utils.SaveScoutPki(args.Hostname, args.PubKey)
		if err != nil {
			return err
		}
		if utils.Config().Scout.AutoAccept {
			data, aptErr := scout.AcceptScout(ScoutInfo)
			if aptErr != nil {
				log.Print(aptErr)
				return aptErr
			}
			reply.Data = data
			reply.Plugins = utils.Plugins
			reply.IgnoreCommands = utils.Config().IgnoreCommand.Commands
			ScoutInfo.Status = "accepted"
		}
	} else {
		// 信任过该HostName的Scout
		ScoutPubKey, is_have := utils.ReadScoutPubKey(args.Hostname)
		if is_have {
			// 存在该Scout的公钥
			if ScoutPubKey == args.PubKey {
				// 公钥未改变
				data, err := scout.AcceptScout(ScoutInfo)
				if err != nil {
					log.Print(err)
					return err
				}
				reply.Data = data
				reply.Plugins = utils.Plugins
				reply.IgnoreCommands = utils.Config().IgnoreCommand.Commands
				ScoutInfo.Status = "accepted"
			} else {
				// 公钥已改变，拒绝
				ScoutInfo.Status = "denied"
			}
		} else {
			// 不存在该Scout的公钥或读取失败，拒绝
			ScoutInfo.Status = "denied"
		}
	}

	reply.Status = ScoutInfo.Status

	if reply.Status != "accepted" {
		cache.Scouts.Put(ScoutInfo)
	}

	return nil
}
