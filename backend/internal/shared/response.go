package shared

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

func SuccessResponseWithMeta(data interface{}, meta interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

func SuccessMessageResponse(message string) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
	}
}

func ErrorResponse(message string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
	}
}

// Pagination meta
type PaginationMeta struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"total_page"`
}

func NewPaginationMeta(page, pageSize int, total int64) PaginationMeta {
	totalPage := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPage++
	}
	return PaginationMeta{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPage,
	}
}

// Helper to parse pagination from query
func ParsePagination(c *fiber.Ctx) (page int, pageSize int) {
	page = c.QueryInt("page", 1)
	pageSize = c.QueryInt("page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return
}
