package dpduk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateDomesticShipment creates a UK/IE/CI outbound shipment.
func (c *Client) CreateDomesticShipment(ctx context.Context, cr Creds, req *CreateDomesticRequest) (*CreateDomesticResult, error) {
	if req == nil {
		return nil, invalidParam("create request is required")
	}
	if strings.TrimSpace(req.NetworkCode) == "" {
		return nil, invalidParam("networkCode is required")
	}
	if strings.TrimSpace(req.ShipmentDate) == "" {
		return nil, invalidParam("shipmentDate is required")
	}
	var out CreateDomesticResult
	if err := c.authedJSON(ctx, cr, http.MethodPost, PathCreateDomestic, req, &out); err != nil {
		return nil, err
	}
	if out.ShipmentID == "" {
		return nil, &APIError{Op: "create", Message: "missing shipmentId"}
	}
	return &out, nil
}

// GetLabel fetches label print strings for a shipment id.
func (c *Client) GetLabel(ctx context.Context, cr Creds, shipmentID string, p LabelParams) (*LabelResult, error) {
	shipmentID = strings.TrimSpace(shipmentID)
	if shipmentID == "" {
		return nil, invalidParam("shipmentId is required")
	}
	if p.PrinterType < 0 || p.PrinterType > 3 {
		return nil, invalidParam("printerType must be 0–3")
	}
	q := url.Values{}
	q.Set("printerType", strconv.Itoa(p.PrinterType))
	if p.PrinterDPI > 0 {
		q.Set("printerDpi", strconv.Itoa(p.PrinterDPI))
	}
	accept := strings.TrimSpace(p.Accept)
	if accept == "" {
		accept = "application/json"
	}
	path := fmt.Sprintf(PathLabelByShipment, url.PathEscape(shipmentID)) + "?" + q.Encode()

	tok, err := c.ensureToken(ctx, cr)
	if err != nil {
		return nil, err
	}
	rawURL := BaseURL(cr.Environment, cr.BaseURL) + path
	headers := map[string]string{
		"Authorization": "Bearer " + tok,
		"Client-Id":     cr.APIKey,
		"Accept":        accept,
	}
	raw, status, err := do(ctx, c.httpClient(), http.MethodGet, rawURL, headers, nil)
	if err != nil {
		return nil, err
	}
	if status == 401 {
		c.InvalidateToken(cr)
		tok, err = c.ensureToken(ctx, cr)
		if err != nil {
			return nil, err
		}
		headers["Authorization"] = "Bearer " + tok
		raw, status, err = do(ctx, c.httpClient(), http.MethodGet, rawURL, headers, nil)
		if err != nil {
			return nil, err
		}
	}
	if status >= 300 {
		var env apiEnvelope
		if json.Unmarshal(raw, &env) == nil && len(env.Error) > 0 {
			return nil, mapAPIErrors(env.Error, raw)
		}
		return nil, &APIError{Op: "get label", StatusCode: status, Body: string(raw)}
	}
	if strings.Contains(strings.ToLower(accept), "json") {
		var out LabelResult
		if err := decodeData(raw, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return &LabelResult{}, nil
	}
	return &LabelResult{PrintString: []string{s}}, nil
}

// ValidateOutboundServices returns available networkKeys for a lane.
func (c *Client) ValidateOutboundServices(ctx context.Context, cr Creds, req *ValidateOutboundRequest) ([]OutboundService, error) {
	if req == nil {
		return nil, invalidParam("validate request is required")
	}
	if req.NumberOfParcels < 1 {
		return nil, invalidParam("numberOfParcels must be >= 1")
	}
	if req.TotalWeight <= 0 {
		return nil, invalidParam("totalWeight must be > 0")
	}
	var out []OutboundService
	if err := c.authedJSON(ctx, cr, http.MethodPost, PathValidateOutbound, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
