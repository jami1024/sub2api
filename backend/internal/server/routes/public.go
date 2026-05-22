package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	pub := v1.Group("/public")
	{
		pub.GET("/announcements", h.Announcement.ListPublic)
	}
}
