package postio

// Models for every Postio API response shape. Field names match the
// OpenAPI spec exactly (camelCase JSON tags) so callers can serialize
// envelopes back out without round-trip drift. Generated equivalent
// would live in this file; kept hand-written because the surface is
// 6 endpoints / 15 schemas — see CLAUDE.md.

// Performance is the per-request timing breakdown returned with every
// successful response.
type Performance struct {
	WorkerMs int `json:"workerMs"`
	LookupMs int `json:"lookupMs"`
}

// Meta is the response envelope metadata — request_id, count, timing.
type Meta struct {
	CountResults int         `json:"countResults"`
	RequestID    string      `json:"requestId"`
	Performance  Performance `json:"performance"`
}

// MetaConnect is the meta block for /connect (no count field).
type MetaConnect struct {
	RequestID   string      `json:"requestId"`
	Performance Performance `json:"performance"`
}

// AddressSearchResult is a single typeahead hit.
type AddressSearchResult struct {
	UDPRN      int    `json:"udprn"`
	Suggestion string `json:"suggestion"`
}

// Address is a full UK address record from Royal Mail PAF +
// Ordnance Survey. Most fields are optional and may be empty strings.
type Address struct {
	UDPRN                   int     `json:"udprn"`
	Postcode                string  `json:"postcode"`
	PostcodeOutward         *string `json:"postcode_outward,omitempty"`
	PostcodeInward          *string `json:"postcode_inward,omitempty"`
	PostcodeType            *string `json:"postcode_type,omitempty"`
	AddressLine1            *string `json:"address_line_1,omitempty"`
	AddressLine2            *string `json:"address_line_2,omitempty"`
	AddressLine3            *string `json:"address_line_3,omitempty"`
	PostTown                *string `json:"post_town,omitempty"`
	OrganisationName        *string `json:"organisation_name,omitempty"`
	DepartmentName          *string `json:"department_name,omitempty"`
	BuildingName            *string `json:"building_name,omitempty"`
	BuildingNumber          *string `json:"building_number,omitempty"`
	SubBuildingName         *string `json:"sub_building_name,omitempty"`
	POBox                   *string `json:"po_box,omitempty"`
	Thoroughfare            *string `json:"thoroughfare,omitempty"`
	DependentThoroughfare   *string `json:"dependent_thoroughfare,omitempty"`
	DependentLocality       *string `json:"dependent_locality,omitempty"`
	DoubleDependentLocality *string `json:"double_dependent_locality,omitempty"`
	DeliveryPointSuffix     *string `json:"delivery_point_suffix,omitempty"`
	Country                 *string `json:"country,omitempty"`
	County                  *string `json:"county,omitempty"`
	District                *string `json:"district,omitempty"`
	Ward                    *string `json:"ward,omitempty"`
	Latitude                *float64 `json:"latitude,omitempty"`
	Longitude               *float64 `json:"longitude,omitempty"`
	Eastings                *int     `json:"eastings,omitempty"`
	Northings               *int     `json:"northings,omitempty"`
}

// EmailResult is the validation verdict for a single email address.
type EmailResult struct {
	Email          string  `json:"email"`
	IsValidSyntax  bool    `json:"isValidSyntax"`
	DidYouMean     *string `json:"didYouMean"`
	IsDisposable   bool    `json:"isDisposable"`
	IsFreeProvider bool    `json:"isFreeProvider"`
	IsRoleAccount  bool    `json:"isRoleAccount"`
	MXFound        bool    `json:"mxFound"`
	SMTPCheck      *string `json:"smtpCheck"`
	IsCatchAll     *bool   `json:"isCatchAll"`
	Deliverability string  `json:"deliverability"`
}

// Deliverability constants — mirror the OpenAPI Deliverability enum.
const (
	DeliverabilityDeliverable   = "deliverable"
	DeliverabilityUndeliverable = "undeliverable"
	DeliverabilityRisky         = "risky"
	DeliverabilityUnknown       = "unknown"
	DeliverabilityInvalid       = "invalid"
)

// PhoneResult is the validation verdict for a single phone number.
//
// SPEC DRIFT (2026-05-02): the OpenAPI spec marks every nullable field
// as `required` with type [string, null], but the live API drops them
// entirely on invalid input. Pointer fields with `omitempty` handle the
// missing case cleanly. Spec also says IsReachable is string|null but
// the live API returns bool — interface{} preserves either.
type PhoneResult struct {
	Number              string      `json:"number"`
	IsValid             bool        `json:"isValid"`
	IsPossible          bool        `json:"isPossible"`
	Type                *string     `json:"type,omitempty"`
	CountryCode         *string     `json:"countryCode,omitempty"`
	CountryName         *string     `json:"countryName,omitempty"`
	NationalFormat      *string     `json:"nationalFormat,omitempty"`
	InternationalFormat *string     `json:"internationalFormat,omitempty"`
	E164Format          *string     `json:"e164Format,omitempty"`
	OriginalCarrier     *string     `json:"originalCarrier,omitempty"`
	CurrentCarrier      *string     `json:"currentCarrier,omitempty"`
	IsPorted            *bool       `json:"isPorted,omitempty"`
	IsReachable         interface{} `json:"isReachable,omitempty"`
	MCC                 *string     `json:"mcc,omitempty"`
	MNC                 *string     `json:"mnc,omitempty"`
	Level               *string     `json:"level,omitempty"`
	LookupError         *string     `json:"lookupError,omitempty"`
}

// AddressSearchEnvelope is the response from /address/search.
type AddressSearchEnvelope struct {
	Success bool                  `json:"success"`
	Results []AddressSearchResult `json:"results"`
	Meta    Meta                  `json:"meta"`
}

// AddressPostcodeEnvelope is the response from /address/postcode/{postcode}.
type AddressPostcodeEnvelope struct {
	Success bool      `json:"success"`
	Results []Address `json:"results"`
	Meta    Meta      `json:"meta"`
}

// AddressUDPRNEnvelope is the response from /address/udprn/{udprn}.
type AddressUDPRNEnvelope struct {
	Success bool      `json:"success"`
	Results []Address `json:"results"`
	Meta    Meta      `json:"meta"`
}

// EmailEnvelope is the response from /email/{address}.
type EmailEnvelope struct {
	Success bool          `json:"success"`
	Results []EmailResult `json:"results"`
	Meta    Meta          `json:"meta"`
}

// PhoneEnvelope is the response from /phone/{number}.
type PhoneEnvelope struct {
	Success bool          `json:"success"`
	Results []PhoneResult `json:"results"`
	Meta    Meta          `json:"meta"`
}

// ConnectSuccess is the response from /connect.
type ConnectSuccess struct {
	Success bool        `json:"success"`
	Meta    MetaConnect `json:"meta"`
}

// ErrorEnvelope is the API's standard error response body. Surfaced via
// the typed error types in errors.go.
type ErrorEnvelope struct {
	Success bool          `json:"success"`
	Error   string        `json:"error"`
	Details *string       `json:"details,omitempty"`
	Results []interface{} `json:"results"`
	Meta    Meta          `json:"meta"`
}
