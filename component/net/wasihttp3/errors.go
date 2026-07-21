package wasihttp3

import (
	"fmt"
	"strings"

	types "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

func headerErrorString(err types.HeaderError) string {
	switch err.Tag() {
	case types.HeaderErrorInvalidSyntax:
		return "invalid syntax"
	case types.HeaderErrorForbidden:
		return "forbidden header"
	case types.HeaderErrorImmutable:
		return "immutable header fields"
	case types.HeaderErrorSizeExceeded:
		return "size exceeded"
	case types.HeaderErrorOther:
		if v := err.Other(); v.IsSome() {
			return v.Some()
		}
		return "unknown header error"
	default:
		return "unknown header error"
	}
}

func errorCodeString(code types.ErrorCode) string {
	switch code.Tag() {
	case types.ErrorCodeDnsTimeout:
		return "DNS timeout"
	case types.ErrorCodeDnsError:
		p := code.DnsError()
		var parts []string
		if p.Rcode.IsSome() {
			parts = append(parts, fmt.Sprintf("rcode=%s", p.Rcode.Some()))
		}
		if p.InfoCode.IsSome() {
			parts = append(parts, fmt.Sprintf("info-code=%d", p.InfoCode.Some()))
		}
		if len(parts) == 0 {
			return "DNS error"
		}
		return "DNS error (" + strings.Join(parts, ", ") + ")"
	case types.ErrorCodeDestinationNotFound:
		return "destination not found"
	case types.ErrorCodeDestinationUnavailable:
		return "destination unavailable"
	case types.ErrorCodeDestinationIpProhibited:
		return "destination IP prohibited"
	case types.ErrorCodeDestinationIpUnroutable:
		return "destination IP unroutable"
	case types.ErrorCodeConnectionRefused:
		return "connection refused"
	case types.ErrorCodeConnectionTerminated:
		return "connection terminated"
	case types.ErrorCodeConnectionTimeout:
		return "connection timeout"
	case types.ErrorCodeConnectionReadTimeout:
		return "connection read timeout"
	case types.ErrorCodeConnectionWriteTimeout:
		return "connection write timeout"
	case types.ErrorCodeConnectionLimitReached:
		return "connection limit reached"
	case types.ErrorCodeTlsProtocolError:
		return "TLS protocol error"
	case types.ErrorCodeTlsCertificateError:
		return "TLS certificate error"
	case types.ErrorCodeTlsAlertReceived:
		p := code.TlsAlertReceived()
		var parts []string
		if p.AlertId.IsSome() {
			parts = append(parts, fmt.Sprintf("alert-id=%d", p.AlertId.Some()))
		}
		if p.AlertMessage.IsSome() {
			parts = append(parts, fmt.Sprintf("alert-message=%s", p.AlertMessage.Some()))
		}
		if len(parts) == 0 {
			return "TLS alert received"
		}
		return "TLS alert received (" + strings.Join(parts, ", ") + ")"
	case types.ErrorCodeHttpRequestDenied:
		return "HTTP request denied"
	case types.ErrorCodeHttpRequestLengthRequired:
		return "HTTP request length required"
	case types.ErrorCodeHttpRequestBodySize:
		if v := code.HttpRequestBodySize(); v.IsSome() {
			return fmt.Sprintf("HTTP request body size: %d", v.Some())
		}
		return "HTTP request body size"
	case types.ErrorCodeHttpRequestMethodInvalid:
		return "HTTP request method invalid"
	case types.ErrorCodeHttpRequestUriInvalid:
		return "HTTP request URI invalid"
	case types.ErrorCodeHttpRequestUriTooLong:
		return "HTTP request URI too long"
	case types.ErrorCodeHttpRequestHeaderSectionSize:
		if v := code.HttpRequestHeaderSectionSize(); v.IsSome() {
			return fmt.Sprintf("HTTP request header section size: %d", v.Some())
		}
		return "HTTP request header section size"
	case types.ErrorCodeHttpRequestHeaderSize:
		if v := code.HttpRequestHeaderSize(); v.IsSome() {
			return "HTTP request header size " + fieldSizeString(v.Some())
		}
		return "HTTP request header size"
	case types.ErrorCodeHttpRequestTrailerSectionSize:
		if v := code.HttpRequestTrailerSectionSize(); v.IsSome() {
			return fmt.Sprintf("HTTP request trailer section size: %d", v.Some())
		}
		return "HTTP request trailer section size"
	case types.ErrorCodeHttpRequestTrailerSize:
		return "HTTP request trailer size " + fieldSizeString(code.HttpRequestTrailerSize())
	case types.ErrorCodeHttpResponseIncomplete:
		return "HTTP response incomplete"
	case types.ErrorCodeHttpResponseHeaderSectionSize:
		if v := code.HttpResponseHeaderSectionSize(); v.IsSome() {
			return fmt.Sprintf("HTTP response header section size: %d", v.Some())
		}
		return "HTTP response header section size"
	case types.ErrorCodeHttpResponseHeaderSize:
		return "HTTP response header size " + fieldSizeString(code.HttpResponseHeaderSize())
	case types.ErrorCodeHttpResponseBodySize:
		if v := code.HttpResponseBodySize(); v.IsSome() {
			return fmt.Sprintf("HTTP response body size: %d", v.Some())
		}
		return "HTTP response body size"
	case types.ErrorCodeHttpResponseTrailerSectionSize:
		if v := code.HttpResponseTrailerSectionSize(); v.IsSome() {
			return fmt.Sprintf("HTTP response trailer section size: %d", v.Some())
		}
		return "HTTP response trailer section size"
	case types.ErrorCodeHttpResponseTrailerSize:
		return "HTTP response trailer size " + fieldSizeString(code.HttpResponseTrailerSize())
	case types.ErrorCodeHttpResponseTransferCoding:
		if v := code.HttpResponseTransferCoding(); v.IsSome() {
			return fmt.Sprintf("HTTP response transfer coding: %s", v.Some())
		}
		return "HTTP response transfer coding"
	case types.ErrorCodeHttpResponseContentCoding:
		if v := code.HttpResponseContentCoding(); v.IsSome() {
			return fmt.Sprintf("HTTP response content coding: %s", v.Some())
		}
		return "HTTP response content coding"
	case types.ErrorCodeHttpResponseTimeout:
		return "HTTP response timeout"
	case types.ErrorCodeHttpUpgradeFailed:
		return "HTTP upgrade failed"
	case types.ErrorCodeHttpProtocolError:
		return "HTTP protocol error"
	case types.ErrorCodeLoopDetected:
		return "loop detected"
	case types.ErrorCodeConfigurationError:
		return "configuration error"
	case types.ErrorCodeInternalError:
		if v := code.InternalError(); v.IsSome() {
			return "internal error: " + v.Some()
		}
		return "internal error"
	default:
		return fmt.Sprintf("unknown error code: %d", code.Tag())
	}
}

func fieldSizeString(p types.FieldSizePayload) string {
	var parts []string
	if p.FieldName.IsSome() {
		parts = append(parts, fmt.Sprintf("field-name=%s", p.FieldName.Some()))
	}
	if p.FieldSize.IsSome() {
		parts = append(parts, fmt.Sprintf("field-size=%d", p.FieldSize.Some()))
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
