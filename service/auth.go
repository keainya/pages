package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AccountBaseURL 账户系统地址
const AccountBaseURL = "https://account.takemeto.icu"

// AuthRegister 代理注册
func AuthRegister(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	resp, err := http.Post(AccountBaseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(500, Response{Code: 9000, Msg: "账户服务不可用"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	c.Data(resp.StatusCode, "application/json", respBody)
}

// AuthLogin 代理登录，成功后在本服务设置 Session
func AuthLogin(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)

	resp, err := http.Post(AccountBaseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(500, Response{Code: 9000, Msg: "账户服务不可用"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int64                  `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(respBody, &result)

	// 登录成功，将用户信息存入本服务 Session
	if result.Code == 0 && result.Data != nil {
		session := sessions.Default(c)
		session.Set("user_id", result.Data["id"])
		session.Set("username", result.Data["username"])
		session.Set("email", result.Data["email"])
		session.Set("role", result.Data["role"])
		session.Save()
	}

	c.Data(resp.StatusCode, "application/json", respBody)
}

// AuthLogout 登出，清除本服务 Session
func AuthLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(200, Response{Code: 0, Msg: "已登出"})
}

// AuthMe 获取当前登录用户（从本服务 Session 读取）
func AuthMe(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")

	if userID == nil || username == nil {
		c.JSON(200, Response{Code: 2001, Msg: "未登录"})
		return
	}

	c.JSON(200, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{
			"id":       userID,
			"username": username,
			"email":    session.Get("email"),
			"role":     session.Get("role"),
		},
	})
}
