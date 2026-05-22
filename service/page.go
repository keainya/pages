package service

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/keainya/pages/object"
)

// ListPages 获取页面列表
func ListPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	object.Database.Model(&object.Page{}).Count(&total)

	var pages []object.Page
	object.Database.Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&pages)

	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"items":     pages,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetPage 获取单个页面
func GetPage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, Response{Code: -1, Msg: "无效的 ID"})
		return
	}

	var page object.Page
	if err := object.Database.First(&page, id).Error; err != nil {
		c.JSON(404, Response{Code: -1, Msg: "页面不存在"})
		return
	}

	c.JSON(200, Response{Code: 0, Msg: "ok", Data: page})
}

// CreatePage 新建页面
func CreatePage(c *gin.Context) {
	var req struct {
		URLPath     string `json:"url_path" binding:"required"`
		HTMLContent string `json:"html_content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, Response{Code: -1, Msg: "url_path 为必填项"})
		return
	}

	// 去除前后 / 和空格
	req.URLPath = strings.Trim(req.URLPath, "/ ")
	if req.URLPath == "" {
		c.JSON(400, Response{Code: -1, Msg: "url_path 不能为空"})
		return
	}

	// 检查路径是否已存在
	var existing object.Page
	if object.Database.Where("url_path = ?", req.URLPath).First(&existing).Error == nil {
		c.JSON(409, Response{Code: -1, Msg: "该路径已存在"})
		return
	}

	page := object.Page{
		URLPath:     req.URLPath,
		HTMLContent: req.HTMLContent,
	}
	if err := object.Database.Create(&page).Error; err != nil {
		c.JSON(500, Response{Code: -1, Msg: "创建失败"})
		return
	}

	c.JSON(201, Response{Code: 0, Msg: "页面已创建", Data: page})
}

// UpdatePage 更新页面
func UpdatePage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, Response{Code: -1, Msg: "无效的 ID"})
		return
	}

	var page object.Page
	if err := object.Database.First(&page, id).Error; err != nil {
		c.JSON(404, Response{Code: -1, Msg: "页面不存在"})
		return
	}

	var req struct {
		URLPath     string `json:"url_path"`
		HTMLContent string `json:"html_content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, Response{Code: -1, Msg: "参数错误"})
		return
	}

	if req.URLPath != "" {
		req.URLPath = strings.Trim(req.URLPath, "/ ")
		if req.URLPath == "" {
			c.JSON(400, Response{Code: -1, Msg: "url_path 不能为空"})
			return
		}
		// 检查新路径是否被其他页面占用
		var dup object.Page
		if object.Database.Where("url_path = ? AND id != ?", req.URLPath, id).First(&dup).Error == nil {
			c.JSON(409, Response{Code: -1, Msg: "该路径已被其他页面占用"})
			return
		}
		page.URLPath = req.URLPath
	}
	page.HTMLContent = req.HTMLContent

	if err := object.Database.Save(&page).Error; err != nil {
		c.JSON(500, Response{Code: -1, Msg: "更新失败"})
		return
	}

	c.JSON(200, Response{Code: 0, Msg: "页面已更新", Data: page})
}

// DeletePage 删除页面
func DeletePage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, Response{Code: -1, Msg: "无效的 ID"})
		return
	}

	var page object.Page
	if err := object.Database.First(&page, id).Error; err != nil {
		c.JSON(404, Response{Code: -1, Msg: "页面不存在"})
		return
	}

	if err := object.Database.Delete(&page).Error; err != nil {
		c.JSON(500, Response{Code: -1, Msg: "删除失败"})
		return
	}

	c.JSON(200, Response{Code: 0, Msg: "页面已删除"})
}

// ResolvePage 根据 URL 路径解析页面，返回 HTML 内容和是否找到
// 优先级：精确匹配 url_path → url_path + ".html"
// 根路径 "/" → "index" → "index.html"
func ResolvePage(urlPath string) (string, bool) {
	path := strings.TrimPrefix(urlPath, "/")

	// 根路径 → index
	if path == "" {
		path = "index"
	}

	// 尝试精确匹配
	page, found := findPage(path)
	if found {
		return page.HTMLContent, true
	}

	// 若无 .html 后缀，尝试补 .html 再查
	if !strings.HasSuffix(path, ".html") {
		page, found = findPage(path + ".html")
		if found {
			return page.HTMLContent, true
		}
	}

	return "", false
}

// findPage 在数据库中查找指定 url_path 的页面
func findPage(urlPath string) (object.Page, bool) {
	var page object.Page
	if err := object.Database.Where("url_path = ?", urlPath).First(&page).Error; err != nil {
		return page, false
	}
	return page, true
}
