package dpduk

// Creds are DPD UK / DPD Local API credentials.
type Creds struct {
	APIKey      string
	APISecret   string
	BaseURL     string
	Environment string // sandbox | production
}

// Address is the UK-shaped address block.
type Address struct {
	Organisation string `json:"organisation,omitempty"`
	Property     string `json:"property,omitempty"`
	Street       string `json:"street,omitempty"`
	Locality     string `json:"locality,omitempty"`
	Town         string `json:"town,omitempty"`
	County       string `json:"county,omitempty"`
	Postcode     string `json:"postcode,omitempty"`
	CountryCode  string `json:"countryCode,omitempty"`
}

// ContactDetails is name + phone.
type ContactDetails struct {
	ContactName string `json:"contactName,omitempty"`
	Telephone   string `json:"telephone,omitempty"`
}

// NotificationDetails is optional SMS / email notify.
type NotificationDetails struct {
	Email  string `json:"email,omitempty"`
	Mobile string `json:"mobile,omitempty"`
}

// PartyDetails wraps contact + address.
type PartyDetails struct {
	ContactDetails      ContactDetails       `json:"contactDetails"`
	Address             Address              `json:"address"`
	NotificationDetails *NotificationDetails `json:"notificationDetails,omitempty"`
}

// OutboundConsignment is the domestic create consignment body.
type OutboundConsignment struct {
	NetworkCode          string       `json:"networkCode,omitempty"`
	CollectionDetails    PartyDetails `json:"collectionDetails"`
	DeliveryDetails      PartyDetails `json:"deliveryDetails"`
	NumberOfParcels      int          `json:"numberOfParcels"`
	TotalWeight          float64      `json:"totalWeight"`
	ShippingRef1         string       `json:"shippingRef1,omitempty"`
	ShippingRef2         string       `json:"shippingRef2,omitempty"`
	ShippingRef3         string       `json:"shippingRef3,omitempty"`
	DeliveryInstructions string       `json:"deliveryInstructions,omitempty"`
	ParcelDescription    string       `json:"parcelDescription,omitempty"`
	Liability            *bool        `json:"liability,omitempty"`
}

// CreateDomesticRequest creates a UK / IE / CI outbound shipment.
type CreateDomesticRequest struct {
	ShipmentDate        string              `json:"shipmentDate"`
	NetworkCode         string              `json:"networkCode"`
	GenerateCustomsData *bool               `json:"generateCustomsData,omitempty"`
	OutboundConsignment OutboundConsignment `json:"outboundConsignment"`
}

// CreateDomesticResult is data from a successful domestic create.
type CreateDomesticResult struct {
	ShipmentID   string              `json:"shipmentId"`
	Consignments []ConsignmentResult `json:"consignments,omitempty"`
}

// ConsignmentResult is one consignment row under create data.
type ConsignmentResult struct {
	ConsignmentNumber string   `json:"consignmentNumber,omitempty"`
	ParcelNumbers     []string `json:"parcelNumber,omitempty"`
	ParcelIDs         []string `json:"parcelIds,omitempty"`
}

// LabelResult holds print strings from Get Label.
type LabelResult struct {
	PrintString []string `json:"printString,omitempty"`
}

// LabelParams selects printer output for GetLabel.
type LabelParams struct {
	PrinterType int
	PrinterDPI  int
	Accept      string
}

// ValidateOutboundRequest checks available networkKeys for a lane.
type ValidateOutboundRequest struct {
	DeliveryDetails   PartyDetails `json:"deliveryDetails"`
	CollectionDetails PartyDetails `json:"collectionDetails"`
	TotalWeight       float64      `json:"totalWeight"`
	ShipmentType      int          `json:"shipmentType"`
	NumberOfParcels   int          `json:"numberOfParcels"`
}

// OutboundService is one available service from validate outbound.
type OutboundService struct {
	NetworkKey  string `json:"networkKey,omitempty"`
	NetworkDesc string `json:"networkDesc,omitempty"`
	Service     *struct {
		ServiceDesc string `json:"serviceDesc,omitempty"`
		ServiceKey  string `json:"serviceKey,omitempty"`
	} `json:"service,omitempty"`
}

// TokenPair is the auth response.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
}
