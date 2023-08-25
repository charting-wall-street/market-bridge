package throw

import "net/http"

type ExceptionKind = int

const (
	ErrKindUnexpected ExceptionKind = iota
	ErrKindUserError
	ErrKindUnavailable
	ErrKindNotImplemented
	ErrKindNotFound
)

type exceptionStruct struct {
	Message string        `json:"message"`
	Kind    ExceptionKind `json:"unexpected"`
}

type Exception = *exceptionStruct

func New(err error, kind ExceptionKind) Exception {
	return &exceptionStruct{err.Error(), kind}
}

var ErrUnhandled = &exceptionStruct{"unexpected error", ErrKindUnexpected}
var ErrWIP = &exceptionStruct{"work in progress", ErrKindNotImplemented}
var ErrNotAvailable = &exceptionStruct{"source does not has requested data", ErrKindNotImplemented}
var ErrInvalidSymbol = &exceptionStruct{"invalid symbol formatting", ErrKindUserError}
var ErrInvalidSource = &exceptionStruct{"invalid data source", ErrKindUserError}
var ErrInvalidExchange = &exceptionStruct{"invalid exchange", ErrKindUserError}
var ErrInvalidInterval = &exceptionStruct{"invalid interval", ErrKindUserError}
var ErrIntervalNotSupported = &exceptionStruct{"interval not supported", ErrKindUserError}
var ErrSourceNotSupported = &exceptionStruct{"not supported", ErrKindUserError}
var ErrInvalidFromParameter = &exceptionStruct{"parameter from is required for exchange", ErrKindUserError}

func HttpError(w http.ResponseWriter, e Exception) {
	switch e.Kind {
	case ErrKindUnavailable:
		http.Error(w, e.Message, http.StatusServiceUnavailable)
	case ErrKindUserError:
		http.Error(w, e.Message, http.StatusBadRequest)
	case ErrKindNotImplemented:
		http.Error(w, e.Message, http.StatusNotImplemented)
	case ErrKindNotFound:
		http.Error(w, e.Message, http.StatusMethodNotAllowed)
	default:
		http.Error(w, e.Message, http.StatusInternalServerError)
	}
}

func HttpNotImplemented(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotImplemented)
}
