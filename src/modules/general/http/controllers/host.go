package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/cache"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/rmq"
	"github.com/troopstack/troop/src/modules/general/scout"

	"github.com/gin-gonic/gin"
)

func HostList(c *gin.Context) {
	h := gin.H{
		"result": []*model.Host{},
		"error":  "",
		"code":   0,
	}

	hosts := []*model.Host{}
	query := database.DBConn().Preload("Tags")

	// 分页
	paging, query := Pagination(c, query)

	// 执行查询
	query.Find(&hosts)
	if paging {
		var count int
		database.DBConn().Model(&hosts).Count(&count)
		pageRes := model.PageResponseBody{
			Count: count,
			Data:  hosts,
		}
		h["result"] = pageRes
	} else {
		h["result"] = hosts
	}
	c.JSON(http.StatusOK, h)
	return
}

type AllHosts struct {
	Accepted   []*model.Host
	Unaccepted []string
	Denied     []string
}

func AllHostList(c *gin.Context) {
	AllHostsMap := AllHosts{}

	h := gin.H{
		"result": AllHosts{
			Accepted:   []*model.Host{},
			Unaccepted: []string{},
			Denied:     []string{},
		},
		"error": "",
		"code":  0,
	}

	hosts := []*model.Host{}
	database.DBConn().Preload("Tags").Find(&hosts)

	AllHostsMap.Accepted = hosts
	CacheHosts := cache.Scouts.All()
	for i := range CacheHosts {
		if CacheHosts[i].Status == "unaccepted" {
			AllHostsMap.Unaccepted = append(AllHostsMap.Unaccepted, CacheHosts[i].Hostname)
		} else if CacheHosts[i].Status == "denied" {
			AllHostsMap.Denied = append(AllHostsMap.Denied, CacheHosts[i].Hostname)
		}
	}
	h["result"] = AllHostsMap
	c.JSON(http.StatusOK, h)
	return
}

type AllHostKeys struct {
	Accepted   []string
	Unaccepted []string
	Denied     []string
}

func HostKeyList(c *gin.Context) {
	AllHostKeys := AllHostKeys{}
	hosts := []*model.Host{}
	database.DBConn().Preload("Tags").Find(&hosts)
	for i := range hosts {
		AllHostKeys.Accepted = append(AllHostKeys.Accepted, hosts[i].Hostname)
	}
	CacheHosts := cache.Scouts.All()
	for i := range CacheHosts {
		if CacheHosts[i].Status == "unaccepted" {
			AllHostKeys.Unaccepted = append(AllHostKeys.Unaccepted, CacheHosts[i].Hostname)
		} else if CacheHosts[i].Status == "denied" {
			AllHostKeys.Denied = append(AllHostKeys.Denied, CacheHosts[i].Hostname)
		}
	}
	c.JSON(http.StatusOK, AllHostKeys)
	return
}

type ScoutTrust struct {
	Hostname string `json:"hostname" binding:"required"`
}

func AcceptHost(c *gin.Context) {
	s := ScoutTrust{}

	// 校验数据
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint(err.Error())})
		return
	}

	// 查询Hostname是否冲突
	if _, exists := database.IsExistScout(s.Hostname); exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The scout already existed"})
		return
	}

	// 检查未接受的Scout缓存列表是否包含此Hostname
	ScoutInfo, exists := cache.Scouts.Get(s.Hostname)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scout not exist"})
		return
	}

	if ScoutInfo.Status != "unaccepted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not allow accept the scout"})
		return
	}

	data, err := scout.AcceptScout(ScoutInfo)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint(err.Error())})
		return
	}

	// 推送密钥给scout
	ScoutRouteKey := fmt.Sprintf("scout.%s.%s", ScoutInfo.Type, s.Hostname)

	ScoutMessage := model.ScoutMessage{
		Type: "accept",
		Data: data,
	}

	_, err = rmq.AmqpServer.PutIntoQueue("scout", ScoutRouteKey, ScoutMessage)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint(err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "Accept「" + s.Hostname + "」successfully"})
	return
}

func AcceptAllHost(c *gin.Context) {
	UnacceptedHosts := cache.Scouts.UnacceptedHosts()
	if len(UnacceptedHosts) == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "Not match any unaccepted scout"})
		return
	}
	var accepted []string
	for i := range UnacceptedHosts {
		// 查询Hostname是否冲突
		if _, exists := database.IsExistScout(UnacceptedHosts[i].Hostname); exists {
			continue
		}
		data, err := scout.AcceptScout(UnacceptedHosts[i])

		if err != nil {
			continue
		}

		// 推送密钥给scout
		ScoutRouteKey := fmt.Sprintf("scout.%s.%s", UnacceptedHosts[i].Type, UnacceptedHosts[i].Hostname)

		ScoutMessage := model.ScoutMessage{
			Type: "accept",
			Data: data,
		}

		_, err = rmq.AmqpServer.PutIntoQueue("scout", ScoutRouteKey, ScoutMessage)

		if err != nil {
			continue
		}
		accepted = append(accepted, UnacceptedHosts[i].Hostname)
	}
	result := strings.Join(accepted, ",")
	c.JSON(http.StatusOK, gin.H{"result": "Accept「" + result + "」successfully"})
	return
}

func RejectHost(c *gin.Context) {
	s := ScoutTrust{}

	// 校验数据
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint(err.Error())})
		return
	}
	// 检查未接受的Scout缓存列表是否包含此Hostname
	ScoutInfo, exists := cache.Scouts.Get(s.Hostname)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scout not exist"})
		return
	}

	if ScoutInfo.Status != "unaccepted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not allow accept the scout"})
		return
	}

	ScoutInfo.Status = "denied"
	c.JSON(http.StatusOK, gin.H{"result": "Reject「" + s.Hostname + "」successfully"})
	return
}

func RejectAllHost(c *gin.Context) {
	UnacceptedHosts := cache.Scouts.UnacceptedHosts()
	if len(UnacceptedHosts) == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "Not match any unaccepted scout"})
		return
	}
	var rejected []string
	for i := range UnacceptedHosts {
		UnacceptedHosts[i].Status = "denied"
		rejected = append(rejected, UnacceptedHosts[i].Hostname)
	}
	result := strings.Join(rejected, ",")
	c.JSON(http.StatusOK, gin.H{"result": "Reject「" + result + "」successfully"})
	return
}

func DeleteHost(c *gin.Context) {
	s := ScoutTrust{}

	// 校验数据
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint(err.Error())})
		return
	}
	// 检查数据库中是否存在此scout
	dbScout, dbExists := database.IsExistScout(s.Hostname)
	// 检查缓存列表中是否存在此scout
	_, cacheExists := cache.Scouts.Get(s.Hostname)

	if !dbExists && !cacheExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scout not exist"})
		return
	}
	if dbExists {
		database.DBConn().Delete(dbScout)
	}

	if cacheExists {
		cache.Scouts.Delete(s.Hostname)
	}
	c.JSON(http.StatusOK, gin.H{"result": "Delete「" + s.Hostname + "」successfully"})
	return
}

func DeleteAllHost(c *gin.Context) {
	// 检查缓存列表中是否存在scout
	hosts := cache.Scouts.Keys()
	// 检查数据库中是否存在scout
	dbScouts := database.HaveScout()

	if len(hosts) == 0 && len(dbScouts) == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "Not match any unaccepted scout"})
		return
	}

	// 清空缓存的scout
	if len(hosts) > 0 {
		for i := range hosts {
			cache.Scouts.Delete(hosts[i])
		}
	}

	// 清空数据库中的scout
	if len(dbScouts) > 0 {
		for i := range dbScouts {
			database.DBConn().Model(&dbScouts[i]).Association("Tags").Clear()
		}
		database.DBConn().Unscoped().Delete(model.Host{})
		database.DBConn().Unscoped().Delete(model.Tag{})
	}

	c.JSON(http.StatusOK, gin.H{"result": "Delete all scout successfully"})
	return
}
