package rpc

import (
	"errors"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/database"
)

func (t *Scout) UpdateScoutHavePlugins(args *model.UpdateScoutHavePluginRequest, reply *model.SimpleRpcResponse) error {
	if args.Hostname == "" {
		reply.Code = 1
		return errors.New("`Hostname` parameter is missing")
	}

	// 检查DB是否已存在此HostName
	_, dbExists := database.IsExistScout(args.Hostname)

	if !dbExists {
		reply.Code = 1
		return errors.New("host not exists")
	}
	hostQ := model.Host{
		Hostname: args.Hostname,
	}
	database.UpdateSpecifyFieldScout(hostQ, map[string]interface{}{"plugins": strings.Join(args.Plugins, ",")})
	reply.Code = 0
	return nil
}
