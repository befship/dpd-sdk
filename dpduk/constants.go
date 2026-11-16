package dpduk

const (
	EnvSandbox    = "sandbox"
	EnvProduction = "production"

	PathAuthToken        = "/auth/token"
	PathCreateDomestic   = "/v1/customer/shipping/shipments/domestic"
	PathLabelByShipment  = "/v1/customer/shipping/shipments/%s/labels"
	PathValidateOutbound = "/v1/customer/shipping/reference/outboundservices"

	HostProduction = "https://api.dpd.co.uk"
	HostSandbox    = "https://api.dpd.co.uk/sandbox"

	DefaultTokenTTLSeconds = 23 * 60 * 60
)

const (
	PrinterHTML    = 0
	PrinterEPL     = 1
	PrinterCitizen = 2
	PrinterZPL     = 3
)
