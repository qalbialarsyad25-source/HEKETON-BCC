package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	apperrors "main/internal/shared/errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	playground "github.com/go-playground/validator/v10"
)

var setupOnce sync.Once

// embeddedName names embedded (anonymous) struct fields, so fieldPath can drop them:
// the client sends `limit`, not `Paging.limit`. A NUL byte can't clash with a real tag name.
const embeddedName = "\x00"

// Setup makes validation errors name fields the way the client sent them (json, form or uri tag).
// Call it once before serving requests; SetupRouter does.
func Setup() {
	setupOnce.Do(func() {
		engine, ok := binding.Validator.Engine().(*playground.Validate)
		if !ok {
			return
		}

		engine.RegisterTagNameFunc(func(field reflect.StructField) string {
			for _, tag := range []string{"json", "form", "uri"} {
				name := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]
				// "-" only means "not from this source", so the next tag still names the field.
				if name != "" && name != "-" {
					return name
				}
			}
			if field.Anonymous {
				return embeddedName
			}
			return field.Name
		})
	})
}

// BindJSON binds and validates the JSON body. The error is an *AppError, ready for c.Error.
func BindJSON[T any](c *gin.Context) (*T, error) {
	var req T

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		var sliceErrs binding.SliceValidationError
		if errors.As(err, &sliceErrs) {
			if issues := sliceIssues("", reflect.ValueOf(req)); len(issues) > 0 {
				return nil, apperrors.Validation(issues)
			}
		}

		appErr := FromBindError(err)
		// Every decode failure of a body is a malformed body, also ones FromBindError
		// can't recognise (e.g. a bad uuid.UUID or time.Time string).
		if errors.Is(appErr, apperrors.ErrInvalidRequest) {
			return nil, apperrors.ErrMalformedBody.Wrap(err)
		}
		return nil, appErr
	}

	return &req, nil
}

// BindURI binds and validates the path parameters.
func BindURI[T any](c *gin.Context) (*T, error) {
	var req T

	if err := c.ShouldBindUri(&req); err != nil {
		return nil, FromBindError(err)
	}

	return &req, nil
}

// BindQuery binds and validates the query string.
func BindQuery[T any](c *gin.Context) (*T, error) {
	var req T

	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, FromBindError(err)
	}

	return &req, nil
}

// FromBindError turns a Gin binding error into an *AppError:
// failed rules become VALIDATION_FAILED with one issue per field, an unreadable JSON body
// becomes MALFORMED_BODY, a body over the http.MaxBytesReader limit becomes 413 PAYLOAD_TOO_LARGE,
// and anything else (e.g. ?page=abc) becomes INVALID_REQUEST.
func FromBindError(err error) *apperrors.AppError {
	var validationErrs playground.ValidationErrors
	if errors.As(err, &validationErrs) {
		return apperrors.Validation(toIssues("", validationErrs))
	}

	// A top-level slice body ([]T). Gin keeps only the failing items, so their index is lost here;
	// BindJSON re-validates the decoded slice to name issues "[i].field".
	var sliceErrs binding.SliceValidationError
	if errors.As(err, &sliceErrs) {
		var issues []apperrors.Issue
		for _, itemErr := range sliceErrs {
			var itemValidationErrs playground.ValidationErrors
			if errors.As(itemErr, &itemValidationErrs) {
				issues = append(issues, toIssues("", itemValidationErrs)...)
			}
		}
		if len(issues) > 0 {
			return apperrors.Validation(issues)
		}
	}

	// A nil or non-struct target is a bug in the handler, not in the request.
	var invalidValidation *playground.InvalidValidationError
	if errors.As(err, &invalidValidation) {
		return apperrors.Internal(err)
	}

	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return apperrors.ErrPayloadTooLarge.Wrap(err)
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return apperrors.ErrMalformedBody.Wrap(err)
	}

	return apperrors.ErrInvalidRequest.Wrap(err)
}

// sliceIssues re-validates a decoded slice body item by item, so issues keep the item's
// real index ("[2].name", "[1][0].name"). Pointers are followed; nil items are skipped.
func sliceIssues(prefix string, value reflect.Value) []apperrors.Issue {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		err := binding.Validator.ValidateStruct(value.Interface())

		var validationErrs playground.ValidationErrors
		if errors.As(err, &validationErrs) {
			return toIssues(prefix, validationErrs)
		}
		return nil
	}

	var issues []apperrors.Issue
	for i := range value.Len() {
		issues = append(issues, sliceIssues(fmt.Sprintf("%s[%d]", prefix, i), value.Index(i))...)
	}
	return issues
}

func toIssues(prefix string, validationErrs playground.ValidationErrors) []apperrors.Issue {
	issues := make([]apperrors.Issue, 0, len(validationErrs))
	for _, e := range validationErrs {
		field := fieldPath(e)
		if prefix != "" {
			field = prefix + "." + field
		}
		issues = append(issues, apperrors.Issue{
			Field:   field,
			Message: message(e),
		})
	}
	return issues
}

// fieldPath is the field's path without the root struct name and embedded structs, e.g. "items[0].name".
func fieldPath(e playground.FieldError) string {
	segments := splitNamespace(e.Namespace())
	if len(segments) < 2 {
		return e.Field()
	}

	path := make([]string, 0, len(segments)-1)
	for _, segment := range segments[1:] {
		if segment = strings.TrimPrefix(segment, embeddedName); segment != "" {
			path = append(path, segment)
		}
	}
	if len(path) == 0 {
		// The rule failed on the embedded struct itself (e.g. a required *Paging).
		return e.StructField()
	}
	return strings.Join(path, ".")
}

// splitNamespace splits at dots outside brackets, so a generic root type
// ("Envelope[main/internal/x.Item]") or a map key ("tags[a.b]") stays one segment.
func splitNamespace(namespace string) []string {
	var segments []string
	depth, start := 0, 0
	for i := 0; i < len(namespace); i++ {
		switch namespace[i] {
		case '[':
			depth++
		case ']':
			depth--
		case '.':
			if depth == 0 {
				segments = append(segments, namespace[start:i])
				start = i + 1
			}
		}
	}
	return append(segments, namespace[start:])
}

func message(e playground.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must %s at least %s%s", sizeVerb(e), e.Param(), sizeUnit(e))
	case "max":
		return fmt.Sprintf("must %s at most %s%s", sizeVerb(e), e.Param(), sizeUnit(e))
	case "len":
		return fmt.Sprintf("must %s exactly %s%s", sizeVerb(e), e.Param(), sizeUnit(e))
	case "numeric":
		return "must contain only digits"
	case "oneof":
		return fmt.Sprintf("must be one of %s", e.Param())
	case "uuid":
		return "must be a valid uuid"
	default:
		return "is invalid"
	}
}

// min, max and len mean characters for strings, items for slices and maps, and the value for numbers.
func sizeVerb(e playground.FieldError) string {
	switch e.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return "contain"
	default:
		return "be"
	}
}

func sizeUnit(e playground.FieldError) string {
	switch e.Kind() {
	case reflect.String:
		return " characters"
	case reflect.Slice, reflect.Array, reflect.Map:
		if e.Param() == "1" {
			return " item"
		}
		return " items"
	default:
		return ""
	}
}
