package httpServer

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type HttpMethod string
type HttpStatus = int
type HttpContext = fiber.Ctx

const (
	Get    HttpMethod = http.MethodGet
	Post              = http.MethodPost
	Put               = http.MethodPut
	Delete            = http.MethodDelete
)

const (
	Ok                          = http.StatusOK
	PermissionDenied HttpStatus = http.StatusForbidden
	BadRequest                  = http.StatusBadRequest
	InternalError               = http.StatusInternalServerError
)

type Route struct {
	Method  HttpMethod
	Path    string
	Handler func(*HttpContext) error
}

type Request struct {
}

func (r *Request) Normalize() *Request {
	return r
}

var validate = validator.New()

func ValidateStruct(str interface{}) []*FieldError {
	var errors []*FieldError
	err := validate.Struct(str)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element FieldError
			element.FieldName = err.StructNamespace()
			element.CurrentValue = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

type Response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Meta   struct {
		Pagination struct {
			CurrentPage int `json:"current_page"`
			TotalCount  int `json:"total_count"`
			TotalPages  int `json:"total_pages"`
			PageSize    int `json:"page_size"`
		} `json:"pagination"`
		Included interface{} `json:"included"`
	}
}

type ErrorResponse struct {
	Status  bool          `json:"status"` // allways false
	Message string        `json:"message"`
	Code    int           `json:"code"`
	TrackId int           `json:"track_id"`
	Errors  []*FieldError `json:"errors"`
}

type FieldError struct {
	FieldName    string   `json:"field_name"`
	CurrentValue string   `json:"current_value"`
	Errors       []string `json:"errors"`
}

type HttpServer interface {
	Listen() error
	SetRouteGroups(groupName string, routes []Route)
	ActiveApm()
	ActiveSentry()
}
