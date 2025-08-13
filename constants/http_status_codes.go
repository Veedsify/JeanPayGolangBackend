package constants

// HTTP Status Codes
const (
	// 1xx Informational
	StatusContinue           = 100
	StatusSwitchingProtocols = 101
	StatusProcessing         = 102

	// 2xx Success
	StatusOK                   = 200
	StatusCreated              = 201
	StatusAccepted             = 202
	StatusNonAuthoritativeInfo = 203
	StatusNoContent            = 204
	StatusResetContent         = 205
	StatusPartialContent       = 206
	StatusMultiStatus          = 207
	StatusAlreadyReported      = 208
	StatusIMUsed               = 226

	// 3xx Redirection
	StatusMultipleChoices   = 300
	StatusMovedPermanently  = 301
	StatusFound             = 302
	StatusSeeOther          = 303
	StatusNotModified       = 304
	StatusUseProxy          = 305
	StatusSwitchProxy       = 306
	StatusTemporaryRedirect = 307
	StatusPermanentRedirect = 308

	// 4xx Client Error
	StatusBadRequest                  = 400
	StatusUnauthorized                = 401
	StatusPaymentRequired             = 402
	StatusForbidden                   = 403
	StatusNotFound                    = 404
	StatusMethodNotAllowed            = 405
	StatusNotAcceptable               = 406
	StatusProxyAuthRequired           = 407
	StatusRequestTimeout              = 408
	StatusConflict                    = 409
	StatusGone                        = 410
	StatusLengthRequired              = 411
	StatusPreconditionFailed          = 412
	StatusPayloadTooLarge             = 413
	StatusURITooLong                  = 414
	StatusUnsupportedMediaType        = 415
	StatusRangeNotSatisfiable         = 416
	StatusExpectationFailed           = 417
	StatusTeapot                      = 418
	StatusMisdirectedRequest          = 421
	StatusUnprocessableEntity         = 422
	StatusLocked                      = 423
	StatusFailedDependency            = 424
	StatusUpgradeRequired             = 426
	StatusPreconditionRequired        = 428
	StatusTooManyRequests             = 429
	StatusRequestHeaderFieldsTooLarge = 431
	StatusUnavailableForLegalReasons  = 451

	// 5xx Server Error
	StatusInternalServerError           = 500
	StatusNotImplemented                = 501
	StatusBadGateway                    = 502
	StatusServiceUnavailable            = 503
	StatusGatewayTimeout                = 504
	StatusHTTPVersionNotSupported       = 505
	StatusVariantAlsoNegotiates         = 506
	StatusInsufficientStorage           = 507
	StatusLoopDetected                  = 508
	StatusNotExtended                   = 510
	StatusNetworkAuthenticationRequired = 511
)

// GetStatusText returns the text description for an HTTP status code
func GetStatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusAccepted:
		return "Accepted"
	case StatusNoContent:
		return "No Content"
	case StatusBadRequest:
		return "Bad Request"
	case StatusUnauthorized:
		return "Unauthorized"
	case StatusForbidden:
		return "Forbidden"
	case StatusNotFound:
		return "Not Found"
	case StatusMethodNotAllowed:
		return "Method Not Allowed"
	case StatusConflict:
		return "Conflict"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusTooManyRequests:
		return "Too Many Requests"
	case StatusInternalServerError:
		return "Internal Server Error"
	case StatusNotImplemented:
		return "Not Implemented"
	case StatusBadGateway:
		return "Bad Gateway"
	case StatusServiceUnavailable:
		return "Service Unavailable"
	case StatusGatewayTimeout:
		return "Gateway Timeout"
	default:
		return "Unknown Status"
	}
}
