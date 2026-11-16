package dpdlv

import eserviss "github.com/befship/dpd-sdk/internal/generated/eserviss"

// Generated client + request/response types re-exported for stable imports.
type (
	ClientWithResponses = eserviss.ClientWithResponses
	RequestEditorFn     = eserviss.RequestEditorFn

	AddressDTO              = eserviss.AddressDTO
	ShipmentCreationDTO     = eserviss.ShipmentCreationDTO
	ShipmentDTO             = eserviss.ShipmentDTO
	ShipmentParcelDTO       = eserviss.ShipmentParcelDTO
	ShipmentServiceDTO      = eserviss.ShipmentServiceDTO
	LabelCreationDTO        = eserviss.LabelCreationDTO
	LabelResponseDTO        = eserviss.LabelResponseDTO
	TokenRequestDTO         = eserviss.TokenRequestDTO
	DeleteShipmentsParams   = eserviss.DeleteShipmentsParams
	GetStatusTrackingParams = eserviss.GetStatusTrackingParams

	LabelCreationDTOLabelFormat = eserviss.LabelCreationDTOLabelFormat
	LabelCreationDTOPaperSize   = eserviss.LabelCreationDTOPaperSize
)

// NewClientWithResponses is the generated constructor (prefer Client.API / Authed).
var NewClientWithResponses = eserviss.NewClientWithResponses
