package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
)

func Pagination(c *gin.Context, query *gorm.DB) (bool, *gorm.DB) {
	page := c.DefaultQuery("page", "")
	pageSize := c.DefaultQuery("page_size", "10")
	paging := false
	if page != "" {
		pageN, _ := strconv.Atoi(page)
		pageSizeN, _ := strconv.Atoi(pageSize)
		query = query.Limit(pageSizeN).Offset((pageN - 1) * pageSizeN)
		paging = true
	}
	return paging, query
}
