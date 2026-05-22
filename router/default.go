package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keainya/pages/service"
)

func InitRouter(webFS embed.FS) *gin.Engine {
	r := gin.Default()

	// ---- CORS 中间件 ----
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ---- Session 中间件 ----
	store := cookie.NewStore([]byte("change-me-to-a-secure-random-key"))
	r.Use(sessions.Sessions("service_session", store))

	// ---- 嵌入式前端静态文件 ----
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}

	// ---- API 路由 ----
	r.GET("/status", service.Status)

	// 认证代理
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", service.AuthRegister)
			auth.POST("/login", service.AuthLogin)
			auth.POST("/logout", service.AuthLogout)
			auth.GET("/me", service.AuthMe)
		}

		// 页面管理（需认证 + 管理员）
		admin := api.Group("/pages")
		admin.Use(AuthRequired(), AdminRequired())
		{
			admin.GET("", service.ListPages)
			admin.GET("/:id", service.GetPage)
			admin.POST("", service.CreatePage)
			admin.PUT("/:id", service.UpdatePage)
			admin.DELETE("/:id", service.DeletePage)
		}
	}

	// ---- 动态页面 + 静态文件兜底 ----
	r.NoRoute(noRouteHandler(staticFS))

	return r
}

// noRouteHandler 兜底处理：优先匹配数据库动态页面 → 静态文件 → SPA index.html
// 路由优先级: DB 动态页面 > 静态资源文件 > SPA 入口 (index.html)
func noRouteHandler(staticFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// GET 请求尝试动态页面匹配
		if c.Request.Method == "GET" {
			// 静态资源后缀，直接走文件服务
			if isStaticAsset(path) {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}

			// 尝试匹配数据库动态页面
			if html, ok := service.ResolvePage(path); ok {
				c.Data(200, "text/html; charset=utf-8", []byte(html))
				return
			}

			// 非 API 路径：SPA fallback，全部返回 index.html
			serveIndexHTML(c, staticFS)
			return
		}

		// 非 GET 请求回退到嵌入式静态文件
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// serveIndexHTML 从嵌入式文件系统提供 index.html
func serveIndexHTML(c *gin.Context, staticFS fs.FS) {
	data, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		c.String(500, "Internal Server Error")
		return
	}
	c.Data(200, "text/html; charset=utf-8", data)
}

// isStaticAsset 判断是否是前端静态资源请求
func isStaticAsset(path string) bool {
	staticExts := []string{".css", ".js", ".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".woff", ".woff2", ".ttf", ".map"}
	lower := strings.ToLower(path)
	for _, ext := range staticExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(200, service.Response{Code: 2001, Msg: "未登录"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminRequired 管理员中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		if role == nil || role.(string) != "admin" {
			c.JSON(200, service.Response{Code: 2002, Msg: "无权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
