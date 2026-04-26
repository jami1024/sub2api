package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type AffiliateHandler struct {
	affiliateService *service.AffiliateService
}

func NewAffiliateHandler(affiliateService *service.AffiliateService) *AffiliateHandler {
	return &AffiliateHandler{affiliateService: affiliateService}
}

func (h *AffiliateHandler) ListWithdrawalRequests(c *gin.Context) {
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	items, err := h.affiliateService.ListWithdrawalRequests(c.Request.Context(), c.Query("status"), limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

type ReviewAffiliateWithdrawalRequest struct {
	AdminNote string `json:"admin_note"`
}

func (h *AffiliateHandler) RejectWithdrawalRequest(c *gin.Context) {
	requestID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req ReviewAffiliateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	item, err := h.affiliateService.RejectWithdrawalRequest(c.Request.Context(), requestID, subject.UserID, req.AdminNote)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AffiliateHandler) MarkWithdrawalPaid(c *gin.Context) {
	requestID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req ReviewAffiliateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	item, err := h.affiliateService.MarkWithdrawalPaid(c.Request.Context(), requestID, subject.UserID, req.AdminNote)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
