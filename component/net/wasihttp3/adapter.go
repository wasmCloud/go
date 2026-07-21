package wasihttp3

import (
	"fmt"
	"net/http"
	"strings"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

// wasiToHTTPRequest converts an incoming wasi:http Request into a
// [net/http.Request]. The wasi request resource is consumed.
func wasiToHTTPRequest(ir *types.Request) (*http.Request, error) {
	defer ir.Drop()

	method, err := methodToString(ir.GetMethod())
	if err != nil {
		return nil, err
	}

	authority := "localhost"
	if auth := ir.GetAuthority(); auth.IsSome() {
		authority = auth.Some()
	}

	pathWithQuery := "/"
	if p := ir.GetPathWithQuery(); p.IsSome() {
		pathWithQuery = p.Some()
	}

	scheme := "http"
	if s := ir.GetScheme(); s.IsSome() && s.Some().Tag() == types.SchemeHttps {
		scheme = "https"
	}

	headers := ir.GetHeaders()
	entries := headers.CopyAll()
	headers.Drop()

	stream, trailers := types.RequestConsumeBody(ir, unitFuture())
	body := newBodyReader(stream, trailers)

	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s%s", scheme, authority, pathWithQuery), body)
	if err != nil {
		body.Close()
		return nil, err
	}

	toHTTPHeader(entries, &req.Header)
	req.Trailer = body.trailer
	req.Host = authority
	req.RequestURI = pathWithQuery

	return req, nil
}

func methodToString(m types.Method) (string, error) {
	switch m.Tag() {
	case types.MethodConnect:
		return http.MethodConnect, nil
	case types.MethodDelete:
		return http.MethodDelete, nil
	case types.MethodGet:
		return http.MethodGet, nil
	case types.MethodHead:
		return http.MethodHead, nil
	case types.MethodOptions:
		return http.MethodOptions, nil
	case types.MethodPatch:
		return http.MethodPatch, nil
	case types.MethodPost:
		return http.MethodPost, nil
	case types.MethodPut:
		return http.MethodPut, nil
	case types.MethodTrace:
		return http.MethodTrace, nil
	case types.MethodOther:
		return m.Other(), fmt.Errorf("unknown http method '%s'", m.Other())
	default:
		return "", fmt.Errorf("failed to convert http method")
	}
}

func toWASIMethod(s string) types.Method {
	switch s {
	case http.MethodConnect:
		return types.MakeMethodConnect()
	case http.MethodDelete:
		return types.MakeMethodDelete()
	case http.MethodGet:
		return types.MakeMethodGet()
	case http.MethodHead:
		return types.MakeMethodHead()
	case http.MethodOptions:
		return types.MakeMethodOptions()
	case http.MethodPatch:
		return types.MakeMethodPatch()
	case http.MethodPost:
		return types.MakeMethodPost()
	case http.MethodPut:
		return types.MakeMethodPut()
	case http.MethodTrace:
		return types.MakeMethodTrace()
	default:
		return types.MakeMethodOther(s)
	}
}

func toHTTPHeader(src []witTypes.Tuple2[string, []uint8], dest *http.Header) {
	for _, pair := range src {
		dest.Add(pair.F0, string(pair.F1))
	}
}

func toWASIHeaders(headers http.Header) (*types.Fields, error) {
	fields := types.MakeFields()

	for key, vals := range headers {
		fieldVals := make([][]uint8, 0, len(vals))
		for _, val := range vals {
			fieldVals = append(fieldVals, []uint8(val))
		}

		if result := fields.Set(key, fieldVals); result.IsErr() {
			fields.Drop()
			return nil, fmt.Errorf("failed to set header %s to [%s]: %s",
				key, strings.Join(vals, ","), headerErrorString(result.Err()))
		}
	}

	return fields, nil
}

// unitFuture returns a pre-resolved future used as the `res` argument of
// consume-body: it signals that we accept the body unconditionally.
func unitFuture() *witTypes.FutureReader[witTypes.Result[witTypes.Unit, types.ErrorCode]] {
	tx, rx := types.MakeFutureResultUnitErrorCode()
	// FutureWriter.Write blocks until the peer reads, so resolve asynchronously.
	go tx.Write(witTypes.Ok[witTypes.Unit, types.ErrorCode](witTypes.Unit{}))
	return rx
}
