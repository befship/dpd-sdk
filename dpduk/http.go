package dpduk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func do(ctx context.Context, hc *http.Client, method, rawURL string, headers map[string]string, payload any) ([]byte, int, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	var body io.Reader
	if payload != nil {
		switch t := payload.(type) {
		case []byte:
			body = bytes.NewReader(t)
		case string:
			body = strings.NewReader(t)
		default:
			raw, err := json.Marshal(t)
			if err != nil {
				return nil, 0, invalidParam("encode request: " + err.Error())
			}
			body = bytes.NewReader(raw)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, 0, invalidParam("build request: " + err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, 0, &APIError{Op: method + " " + redactURL(rawURL), Message: "transport", Err: err}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, &APIError{Op: "read body", StatusCode: resp.StatusCode, Err: err}
	}
	return raw, resp.StatusCode, nil
}

func redactURL(rawURL string) string {
	if i := strings.IndexByte(rawURL, '?'); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}

type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error []apiErrorItem  `json:"error"`
}

type apiErrorItem struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	FieldName string `json:"fieldName"`
}

func decodeData(raw []byte, dest any) error {
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return &APIError{Op: "decode", Message: err.Error(), Body: truncate(string(raw), 400)}
	}
	if len(env.Error) > 0 {
		return mapAPIErrors(env.Error, raw)
	}
	if dest == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Data, dest); err != nil {
		return &APIError{Op: "decode data", Message: err.Error(), Body: truncate(string(raw), 400)}
	}
	return nil
}

func mapAPIErrors(errs []apiErrorItem, raw []byte) error {
	e := errs[0]
	msg := e.Message
	if e.FieldName != "" {
		msg = fmt.Sprintf("%s (%s)", msg, e.FieldName)
	}
	status := e.Status
	if status == 0 {
		status = 400
	}
	code := strings.ToLower(e.ErrorCode)
	auth := status == 401 || status == 403 || strings.Contains(code, "auth/")
	return &APIError{
		Op: "api", StatusCode: status, Code: e.ErrorCode, Message: msg,
		Body: string(raw), Auth: auth,
	}
}
