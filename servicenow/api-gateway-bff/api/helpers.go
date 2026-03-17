package api

import (
	"errors"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(CtxKeyRequestID).(string)
	correlationID, _ := r.Context().Value(CtxKeyCorrelationID).(string)
	WriteError(w, http.StatusMethodNotAllowed, CodeValidationError, "Method not allowed", requestID, correlationID, nil)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, field, issue string) {
	requestID, _ := r.Context().Value(CtxKeyRequestID).(string)
	correlationID, _ := r.Context().Value(CtxKeyCorrelationID).(string)
	WriteError(w, http.StatusBadRequest, CodeValidationError, "Validation failed", requestID, correlationID, []DetailItem{{Field: field, Issue: issue}})
}

func writeDownstreamError(w http.ResponseWriter, r *http.Request, err error) {
	requestID, _ := r.Context().Value(CtxKeyRequestID).(string)
	correlationID, _ := r.Context().Value(CtxKeyCorrelationID).(string)
	code := CodeInternal
	httpStatus := http.StatusInternalServerError
	msg := "Internal error"
	if errors.Is(err, errNotFound) {
		code = CodeNotFound
		httpStatus = http.StatusNotFound
		msg = "Not found"
	} else if errors.Is(err, errValidation) {
		code = CodeValidationError
		httpStatus = http.StatusBadRequest
		msg = err.Error()
	} else if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			code = CodeValidationError
			httpStatus = http.StatusBadRequest
			msg = st.Message()
		case codes.NotFound:
			code = CodeNotFound
			httpStatus = http.StatusNotFound
			msg = st.Message()
		case codes.FailedPrecondition:
			code = CodeValidationError
			httpStatus = http.StatusConflict
			msg = st.Message()
		default:
			msg = st.Message()
		}
	} else if err != nil {
		// Surface actual error for 500 so clients and logs can debug (e.g. connection refused, gRPC desc)
		msg = err.Error()
	}
	WriteError(w, httpStatus, code, msg, requestID, correlationID, nil)
}

var errNotFound = errors.New("not found")
var errValidation = errors.New("validation error")

func parsePagination(r *http.Request) (pageSize int32, pageToken string) {
	pageSize = 50
	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			pageSize = int32(n)
		}
	}
	pageToken = r.URL.Query().Get("page_token")
	return pageSize, pageToken
}
