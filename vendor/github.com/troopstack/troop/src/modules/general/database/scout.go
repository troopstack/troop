package database

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/troopstack/troop/src/model"

	"github.com/chenhg5/collection"
)

func UpdateScout(info *model.ScoutInfo) {
	log.Print(strings.Join(info.Plugins, ","))
	host := model.Host{
		Hostname:     info.Hostname,
		Ip:           info.IP,
		ScoutVersion: info.ScoutVersion,
		Type:         info.Type,
		Status:       "accepted",
		HandshakeAt:  time.Now(),
		AES:          info.AES,
		OS:           info.OS,
		Tags:         []model.Tag{},
		Plugins:      strings.Join(info.Plugins, ","),
	}
	for i := range info.Tags {
		tagName := strings.TrimSpace(info.Tags[i])
		if tagName != "" {
			tag := model.Tag{
				Name: tagName,
			}
			DBConn().FirstOrCreate(&tag, model.Tag{Name: tagName})
			host.Tags = append(host.Tags, tag)
		}
	}
	if hostQ, exists := IsExistScout(info.Hostname); exists {
		hostTagIds := []int{}
		for i := range host.Tags {
			hostTagIds = append(hostTagIds, host.Tags[i].ID)
		}
		hostQTags := []model.Tag{}
		for i := range hostQ.Tags {
			hostQTags = append(hostQTags, hostQ.Tags[i])
		}

		// 删除Host和Tag关系表数据
		for i := range hostQTags {
			if !collection.Collect(hostTagIds).Contains(hostQTags[i].ID) {
				DBConn().Model(&hostQ).Association("Tags").Delete(hostQTags[i])
			}
		}
		DBConn().Model(&hostQ).Update(&host)

		// 当使用struct更新时，FORM将仅更新具有非空值的字段，当插件为空时需要手动更新
		if host.Plugins == "" {
			UpdateSpecifyFieldScout(hostQ, map[string]interface{}{"plugins": ""})
		}
	} else {
		DBConn().Create(&host)
	}
}

func IsExistScout(hostname string) (model.Host, bool) {
	host := model.Host{}
	DBConn().Preload("Tags").Where("hostname = ?", hostname).First(&host)
	if host.ID == 0 {
		return host, false
	} else {
		return host, true
	}
}

func IsExistTypeScout(hostname string, scoutType string) (model.Host, bool) {
	host := model.Host{}
	DBConn().Where("hostname = ? AND type = ?", hostname, scoutType).First(&host)
	if host.ID == 0 {
		return host, false
	} else {
		return host, true
	}
}

func HaveScout() []model.Host {
	host := []model.Host{}
	DBConn().Find(&host)
	return host
}

func IsExistTag(name string) (model.Tag, bool) {
	tag := model.Tag{}
	DBConn().Where("name = ?", name).First(&tag)
	if tag.ID == 0 {
		return tag, false
	} else {
		return tag, true
	}
}

func MatchedScout(t *model.MatchedScout) ([]*model.Host, error) {
	// 当 TargetType = `*` 时，匹配所有Scout类型
	// 当 TargetType为空时，默认匹配给server类型
	// 当 Target = `*`时，匹配所有指定类型或不指定类型的Scout
	scouts := []*model.Host{}
	target := "scout"
	if t.TargetType == "" {
		t.TargetType = "server"
	}
	if t.TargetType != "*" {
		target = target + "." + t.TargetType
	}

	// 检索Scout
	where := []string{}
	if t.Target != "*" {
		scoutMap := strings.Replace(t.Target, " ", "", -1)
		scoutMap = strings.Replace(scoutMap, ",", "','", -1)
		if t.TargetType == "*" {
			where = append(where, fmt.Sprintf("hostname in ('%s')", scoutMap))
		} else {
			where = append(where, fmt.Sprintf("hostname in ('%s')", scoutMap))
			where = append(where, fmt.Sprintf("type = '%s'", t.TargetType))
		}
		if len(scouts) > 0 {
			target = target + "." + t.Target
		}
	}
	if t.OS != "" {
		where = append(where, fmt.Sprintf("os = '%s'", t.OS))
	}
	if t.Tag != "" {
		tag, tagExists := IsExistTag(t.Tag)
		if !tagExists {
			return scouts, errors.New("No scouts matched the target")
		}
		if len(where) == 0 {
			DBConn().Model(&tag).Preload("Tags").Related(&scouts, "Hosts")
		} else {
			DBConn().Model(&tag).Preload("Tags").Where(
				strings.Join(where, " AND ")).Related(&scouts, "Hosts")
		}

	} else {
		if len(where) == 0 {
			DBConn().Preload("Tags").Find(&scouts)
		} else {
			DBConn().Preload("Tags").Where(strings.Join(where, " AND ")).Find(&scouts)
		}
	}
	return scouts, nil
}

func UpdateSpecifyFieldScout(hostQ model.Host, data map[string]interface{}) {
	DBConn().Model(&hostQ).Update(data)
}
