package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/asimovo-platform/metering-client/client/models"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/runtime"
)

const (
	PortalTokenScopes = "portalToken.Scopes"
)

// Event CloudEvents Specification JSON Schema
type Event = event.Event

// IdOrSlug defines model for IdOrSlug.
type IdOrSlug = string

// IngestedEvent defines model for IngestedEvent.
type IngestedEvent struct {
	// Event CloudEvents Specification JSON Schema
	Event           Event   `json:"event"`
	ValidationError *string `json:"validationError,omitempty"`
}

// Meter defines model for Meter.
type Meter = models.Meter

// MeterAggregation The aggregation type to use for the meter.
type MeterAggregation = models.MeterAggregation

// MeterQueryRow defines model for MeterQueryRow.
type MeterQueryRow = models.MeterQueryRow

// PortalToken defines model for PortalToken.
type PortalToken struct {
	AllowedMeterSlugs *[]string  `json:"allowedMeterSlugs,omitempty"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
	Subject           string     `json:"subject"`
	Token             *string    `json:"token,omitempty"`
}

// Problem A Problem Details object (RFC 7807)
type Problem = models.StatusProblem

// WindowSize Aggregation window size.
type WindowSize = models.WindowSize

// MeterIdOrSlug defines model for meterIdOrSlug.
type MeterIdOrSlug = IdOrSlug

// QueryFrom defines model for queryFrom.
type QueryFrom = time.Time

// QueryGroupBy defines model for queryGroupBy.
type QueryGroupBy = []string

// QuerySubject defines model for querySubject.
type QuerySubject = []string

// QueryTo defines model for queryTo.
type QueryTo = time.Time

// QueryWindowSize Aggregation window size.
type QueryWindowSize = WindowSize

// QueryWindowTimeZone defines model for queryWindowTimeZone.
type QueryWindowTimeZone = string

// BadRequestProblemResponse A Problem Details object (RFC 7807)
type BadRequestProblemResponse = Problem

// NotFoundProblemResponse A Problem Details object (RFC 7807)
type NotFoundProblemResponse = Problem

// NotImplementedProblemResponse A Problem Details object (RFC 7807)
type NotImplementedProblemResponse = Problem

// UnauthorizedProblemResponse A Problem Details object (RFC 7807)
type UnauthorizedProblemResponse = Problem

// UnexpectedProblemResponse A Problem Details object (RFC 7807)
type UnexpectedProblemResponse = Problem

// ListEventsParams defines parameters for ListEvents.
type ListEventsParams struct {
	// From Start date-time in RFC 3339 format.
	// Inclusive.
	From *QueryFrom `form:"from,omitempty" json:"from,omitempty"`

	// To End date-time in RFC 3339 format.
	// Inclusive.
	To *QueryTo `form:"to,omitempty" json:"to,omitempty"`

	// Limit Number of events to return.
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`
}

// IngestEventsApplicationCloudeventsBatchPlusJSONBody defines parameters for IngestEvents.
type IngestEventsApplicationCloudeventsBatchPlusJSONBody = []Event

// QueryMeterParams defines parameters for QueryMeter.
type QueryMeterParams struct {
	// From Start date-time in RFC 3339 format.
	// Inclusive.
	From *QueryFrom `form:"from,omitempty" json:"from,omitempty"`

	// To End date-time in RFC 3339 format.
	// Inclusive.
	To *QueryTo `form:"to,omitempty" json:"to,omitempty"`

	// WindowSize If not specified, a single usage aggregate will be returned for the entirety of the specified period for each subject and group.
	WindowSize *QueryWindowSize `form:"windowSize,omitempty" json:"windowSize,omitempty"`

	// WindowTimeZone The value is the name of the time zone as defined in the IANA Time Zone Database (http://www.iana.org/time-zones).
	// If not specified, the UTC timezone will be used.
	WindowTimeZone *QueryWindowTimeZone `form:"windowTimeZone,omitempty" json:"windowTimeZone,omitempty"`
	Subject        *QuerySubject        `form:"subject,omitempty" json:"subject,omitempty"`

	// GroupBy If not specified a single aggregate will be returned for each subject and time window.
	// `subject` is a reserved group by value.
	GroupBy *QueryGroupBy `form:"groupBy,omitempty" json:"groupBy,omitempty"`
}

// QueryPortalMeterParams defines parameters for QueryPortalMeter.
type QueryPortalMeterParams struct {
	// From Start date-time in RFC 3339 format.
	// Inclusive.
	From *QueryFrom `form:"from,omitempty" json:"from,omitempty"`

	// To End date-time in RFC 3339 format.
	// Inclusive.
	To *QueryTo `form:"to,omitempty" json:"to,omitempty"`

	// WindowSize If not specified, a single usage aggregate will be returned for the entirety of the specified period for each subject and group.
	WindowSize *QueryWindowSize `form:"windowSize,omitempty" json:"windowSize,omitempty"`

	// WindowTimeZone The value is the name of the time zone as defined in the IANA Time Zone Database (http://www.iana.org/time-zones).
	// If not specified, the UTC timezone will be used.
	WindowTimeZone *QueryWindowTimeZone `form:"windowTimeZone,omitempty" json:"windowTimeZone,omitempty"`

	// GroupBy If not specified a single aggregate will be returned for each subject and time window.
	// `subject` is a reserved group by value.
	GroupBy *QueryGroupBy `form:"groupBy,omitempty" json:"groupBy,omitempty"`
}

// InvalidatePortalTokensJSONBody defines parameters for InvalidatePortalTokens.
type InvalidatePortalTokensJSONBody struct {
	Subject *string `json:"subject,omitempty"`
}

// IngestEventsApplicationCloudeventsPlusJSONRequestBody defines body for IngestEvents for application/cloudevents+json ContentType.
type IngestEventsApplicationCloudeventsPlusJSONRequestBody = Event

// IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody defines body for IngestEvents for application/cloudevents-batch+json ContentType.
type IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody = IngestEventsApplicationCloudeventsBatchPlusJSONBody

// CreateMeterJSONRequestBody defines body for CreateMeter for application/json ContentType.
type CreateMeterJSONRequestBody = Meter

// CreatePortalTokenJSONRequestBody defines body for CreatePortalToken for application/json ContentType.
type CreatePortalTokenJSONRequestBody = PortalToken

// InvalidatePortalTokensJSONRequestBody defines body for InvalidatePortalTokens for application/json ContentType.
type InvalidatePortalTokensJSONRequestBody InvalidatePortalTokensJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ListEvents request
	ListEvents(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// IngestEventsWithBody request with any body
	IngestEventsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	IngestEventsWithApplicationCloudeventsPlusJSONBody(ctx context.Context, body IngestEventsApplicationCloudeventsPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	IngestEventsWithApplicationCloudeventsBatchPlusJSONBody(ctx context.Context, body IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListMeters request
	ListMeters(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateMeterWithBody request with any body
	CreateMeterWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateMeter(ctx context.Context, body CreateMeterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteMeter request
	DeleteMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetMeter request
	GetMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error)

	// QueryMeter request
	QueryMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, params *QueryMeterParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListMeterSubjects request
	ListMeterSubjects(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error)

	// QueryPortalMeter request
	QueryPortalMeter(ctx context.Context, meterSlug string, params *QueryPortalMeterParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreatePortalTokenWithBody request with any body
	CreatePortalTokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreatePortalToken(ctx context.Context, body CreatePortalTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// InvalidatePortalTokensWithBody request with any body
	InvalidatePortalTokensWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	InvalidatePortalTokens(ctx context.Context, body InvalidatePortalTokensJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ListEvents(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListEventsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) IngestEventsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewIngestEventsRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) IngestEventsWithApplicationCloudeventsPlusJSONBody(ctx context.Context, body IngestEventsApplicationCloudeventsPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewIngestEventsRequestWithApplicationCloudeventsPlusJSONBody(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) IngestEventsWithApplicationCloudeventsBatchPlusJSONBody(ctx context.Context, body IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewIngestEventsRequestWithApplicationCloudeventsBatchPlusJSONBody(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListMeters(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListMetersRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateMeterWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateMeterRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateMeter(ctx context.Context, body CreateMeterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateMeterRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteMeterRequest(c.Server, meterIdOrSlug)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetMeterRequest(c.Server, meterIdOrSlug)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) QueryMeter(ctx context.Context, meterIdOrSlug MeterIdOrSlug, params *QueryMeterParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewQueryMeterRequest(c.Server, meterIdOrSlug, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListMeterSubjects(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListMeterSubjectsRequest(c.Server, meterIdOrSlug)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) QueryPortalMeter(ctx context.Context, meterSlug string, params *QueryPortalMeterParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewQueryPortalMeterRequest(c.Server, meterSlug, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreatePortalTokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreatePortalTokenRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreatePortalToken(ctx context.Context, body CreatePortalTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreatePortalTokenRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) InvalidatePortalTokensWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewInvalidatePortalTokensRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) InvalidatePortalTokens(ctx context.Context, body InvalidatePortalTokensJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewInvalidatePortalTokensRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewListEventsRequest generates requests for ListEvents
func NewListEventsRequest(server string, params *ListEventsParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/events")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.From != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "from", runtime.ParamLocationQuery, *params.From); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.To != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "to", runtime.ParamLocationQuery, *params.To); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.Limit != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "limit", runtime.ParamLocationQuery, *params.Limit); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewIngestEventsRequestWithApplicationCloudeventsPlusJSONBody calls the generic IngestEvents builder with application/cloudevents+json body
func NewIngestEventsRequestWithApplicationCloudeventsPlusJSONBody(server string, body IngestEventsApplicationCloudeventsPlusJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewIngestEventsRequestWithBody(server, "application/cloudevents+json", bodyReader)
}

// NewIngestEventsRequestWithApplicationCloudeventsBatchPlusJSONBody calls the generic IngestEvents builder with application/cloudevents-batch+json body
func NewIngestEventsRequestWithApplicationCloudeventsBatchPlusJSONBody(server string, body IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewIngestEventsRequestWithBody(server, "application/cloudevents-batch+json", bodyReader)
}

// NewIngestEventsRequestWithBody generates requests for IngestEvents with any type of body
func NewIngestEventsRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/events")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListMetersRequest generates requests for ListMeters
func NewListMetersRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateMeterRequest calls the generic CreateMeter builder with application/json body
func NewCreateMeterRequest(server string, body CreateMeterJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateMeterRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateMeterRequestWithBody generates requests for CreateMeter with any type of body
func NewCreateMeterRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteMeterRequest generates requests for DeleteMeter
func NewDeleteMeterRequest(server string, meterIdOrSlug MeterIdOrSlug) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "meterIdOrSlug", runtime.ParamLocationPath, meterIdOrSlug)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetMeterRequest generates requests for GetMeter
func NewGetMeterRequest(server string, meterIdOrSlug MeterIdOrSlug) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "meterIdOrSlug", runtime.ParamLocationPath, meterIdOrSlug)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewQueryMeterRequest generates requests for QueryMeter
func NewQueryMeterRequest(server string, meterIdOrSlug MeterIdOrSlug, params *QueryMeterParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "meterIdOrSlug", runtime.ParamLocationPath, meterIdOrSlug)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters/%s/query", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.From != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "from", runtime.ParamLocationQuery, *params.From); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.To != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "to", runtime.ParamLocationQuery, *params.To); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.WindowSize != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "windowSize", runtime.ParamLocationQuery, *params.WindowSize); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.WindowTimeZone != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "windowTimeZone", runtime.ParamLocationQuery, *params.WindowTimeZone); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.Subject != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "subject", runtime.ParamLocationQuery, *params.Subject); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.GroupBy != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "groupBy", runtime.ParamLocationQuery, *params.GroupBy); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListMeterSubjectsRequest generates requests for ListMeterSubjects
func NewListMeterSubjectsRequest(server string, meterIdOrSlug MeterIdOrSlug) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "meterIdOrSlug", runtime.ParamLocationPath, meterIdOrSlug)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/meters/%s/subjects", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewQueryPortalMeterRequest generates requests for QueryPortalMeter
func NewQueryPortalMeterRequest(server string, meterSlug string, params *QueryPortalMeterParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "meterSlug", runtime.ParamLocationPath, meterSlug)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/portal/meters/%s/query", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.From != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "from", runtime.ParamLocationQuery, *params.From); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.To != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "to", runtime.ParamLocationQuery, *params.To); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.WindowSize != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "windowSize", runtime.ParamLocationQuery, *params.WindowSize); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.WindowTimeZone != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "windowTimeZone", runtime.ParamLocationQuery, *params.WindowTimeZone); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		if params.GroupBy != nil {
			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "groupBy", runtime.ParamLocationQuery, *params.GroupBy); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}
		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreatePortalTokenRequest calls the generic CreatePortalToken builder with application/json body
func NewCreatePortalTokenRequest(server string, body CreatePortalTokenJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreatePortalTokenRequestWithBody(server, "application/json", bodyReader)
}

// NewCreatePortalTokenRequestWithBody generates requests for CreatePortalToken with any type of body
func NewCreatePortalTokenRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/portal/tokens")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewInvalidatePortalTokensRequest calls the generic InvalidatePortalTokens builder with application/json body
func NewInvalidatePortalTokensRequest(server string, body InvalidatePortalTokensJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewInvalidatePortalTokensRequestWithBody(server, "application/json", bodyReader)
}

// NewInvalidatePortalTokensRequestWithBody generates requests for InvalidatePortalTokens with any type of body
func NewInvalidatePortalTokensRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/v1/portal/tokens/invalidate")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ListEventsWithResponse request
	ListEventsWithResponse(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*ListEventsResponse, error)

	// IngestEventsWithBodyWithResponse request with any body
	IngestEventsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error)

	IngestEventsWithApplicationCloudeventsPlusJSONBodyWithResponse(ctx context.Context, body IngestEventsApplicationCloudeventsPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error)

	IngestEventsWithApplicationCloudeventsBatchPlusJSONBodyWithResponse(ctx context.Context, body IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error)

	// ListMetersWithResponse request
	ListMetersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListMetersResponse, error)

	// CreateMeterWithBodyWithResponse request with any body
	CreateMeterWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateMeterResponse, error)

	CreateMeterWithResponse(ctx context.Context, body CreateMeterJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateMeterResponse, error)

	// DeleteMeterWithResponse request
	DeleteMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*DeleteMeterResponse, error)

	// GetMeterWithResponse request
	GetMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*GetMeterResponse, error)

	// QueryMeterWithResponse request
	QueryMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, params *QueryMeterParams, reqEditors ...RequestEditorFn) (*QueryMeterResponse, error)

	// ListMeterSubjectsWithResponse request
	ListMeterSubjectsWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*ListMeterSubjectsResponse, error)

	// QueryPortalMeterWithResponse request
	QueryPortalMeterWithResponse(ctx context.Context, meterSlug string, params *QueryPortalMeterParams, reqEditors ...RequestEditorFn) (*QueryPortalMeterResponse, error)

	// CreatePortalTokenWithBodyWithResponse request with any body
	CreatePortalTokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreatePortalTokenResponse, error)

	CreatePortalTokenWithResponse(ctx context.Context, body CreatePortalTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*CreatePortalTokenResponse, error)

	// InvalidatePortalTokensWithBodyWithResponse request with any body
	InvalidatePortalTokensWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*InvalidatePortalTokensResponse, error)

	InvalidatePortalTokensWithResponse(ctx context.Context, body InvalidatePortalTokensJSONRequestBody, reqEditors ...RequestEditorFn) (*InvalidatePortalTokensResponse, error)
}

type ListEventsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *[]IngestedEvent
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r ListEventsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListEventsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type IngestEventsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r IngestEventsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r IngestEventsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListMetersResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *[]Meter
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r ListMetersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListMetersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateMeterResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON201                       *Meter
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSON501     *NotImplementedProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r CreateMeterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateMeterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteMeterResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSON404     *NotFoundProblemResponse
	ApplicationproblemJSON501     *NotImplementedProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r DeleteMeterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteMeterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetMeterResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *Meter
	ApplicationproblemJSON404     *NotFoundProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r GetMeterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetMeterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type QueryMeterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Data []MeterQueryRow `json:"data"`
		From *time.Time      `json:"from,omitempty"`
		To   *time.Time      `json:"to,omitempty"`

		// WindowSize Aggregation window size.
		WindowSize *WindowSize `json:"windowSize,omitempty"`
	}
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r QueryMeterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r QueryMeterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListMeterSubjectsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *[]string
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r ListMeterSubjectsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListMeterSubjectsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type QueryPortalMeterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Data []MeterQueryRow `json:"data"`
		From *time.Time      `json:"from,omitempty"`
		To   *time.Time      `json:"to,omitempty"`

		// WindowSize Aggregation window size.
		WindowSize *WindowSize `json:"windowSize,omitempty"`
	}
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSON401     *UnauthorizedProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r QueryPortalMeterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r QueryPortalMeterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreatePortalTokenResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *PortalToken
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r CreatePortalTokenResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreatePortalTokenResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type InvalidatePortalTokensResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSON400     *BadRequestProblemResponse
	ApplicationproblemJSONDefault *UnexpectedProblemResponse
}

// Status returns HTTPResponse.Status
func (r InvalidatePortalTokensResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r InvalidatePortalTokensResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ListEventsWithResponse request returning *ListEventsResponse
func (c *ClientWithResponses) ListEventsWithResponse(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*ListEventsResponse, error) {
	rsp, err := c.ListEvents(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListEventsResponse(rsp)
}

// IngestEventsWithBodyWithResponse request with arbitrary body returning *IngestEventsResponse
func (c *ClientWithResponses) IngestEventsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error) {
	rsp, err := c.IngestEventsWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseIngestEventsResponse(rsp)
}

func (c *ClientWithResponses) IngestEventsWithApplicationCloudeventsPlusJSONBodyWithResponse(ctx context.Context, body IngestEventsApplicationCloudeventsPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error) {
	rsp, err := c.IngestEventsWithApplicationCloudeventsPlusJSONBody(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseIngestEventsResponse(rsp)
}

func (c *ClientWithResponses) IngestEventsWithApplicationCloudeventsBatchPlusJSONBodyWithResponse(ctx context.Context, body IngestEventsApplicationCloudeventsBatchPlusJSONRequestBody, reqEditors ...RequestEditorFn) (*IngestEventsResponse, error) {
	rsp, err := c.IngestEventsWithApplicationCloudeventsBatchPlusJSONBody(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseIngestEventsResponse(rsp)
}

// ListMetersWithResponse request returning *ListMetersResponse
func (c *ClientWithResponses) ListMetersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListMetersResponse, error) {
	rsp, err := c.ListMeters(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListMetersResponse(rsp)
}

// CreateMeterWithBodyWithResponse request with arbitrary body returning *CreateMeterResponse
func (c *ClientWithResponses) CreateMeterWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateMeterResponse, error) {
	rsp, err := c.CreateMeterWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateMeterResponse(rsp)
}

func (c *ClientWithResponses) CreateMeterWithResponse(ctx context.Context, body CreateMeterJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateMeterResponse, error) {
	rsp, err := c.CreateMeter(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateMeterResponse(rsp)
}

// DeleteMeterWithResponse request returning *DeleteMeterResponse
func (c *ClientWithResponses) DeleteMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*DeleteMeterResponse, error) {
	rsp, err := c.DeleteMeter(ctx, meterIdOrSlug, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteMeterResponse(rsp)
}

// GetMeterWithResponse request returning *GetMeterResponse
func (c *ClientWithResponses) GetMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*GetMeterResponse, error) {
	rsp, err := c.GetMeter(ctx, meterIdOrSlug, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetMeterResponse(rsp)
}

// QueryMeterWithResponse request returning *QueryMeterResponse
func (c *ClientWithResponses) QueryMeterWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, params *QueryMeterParams, reqEditors ...RequestEditorFn) (*QueryMeterResponse, error) {
	rsp, err := c.QueryMeter(ctx, meterIdOrSlug, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseQueryMeterResponse(rsp)
}

// ListMeterSubjectsWithResponse request returning *ListMeterSubjectsResponse
func (c *ClientWithResponses) ListMeterSubjectsWithResponse(ctx context.Context, meterIdOrSlug MeterIdOrSlug, reqEditors ...RequestEditorFn) (*ListMeterSubjectsResponse, error) {
	rsp, err := c.ListMeterSubjects(ctx, meterIdOrSlug, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListMeterSubjectsResponse(rsp)
}

// QueryPortalMeterWithResponse request returning *QueryPortalMeterResponse
func (c *ClientWithResponses) QueryPortalMeterWithResponse(ctx context.Context, meterSlug string, params *QueryPortalMeterParams, reqEditors ...RequestEditorFn) (*QueryPortalMeterResponse, error) {
	rsp, err := c.QueryPortalMeter(ctx, meterSlug, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseQueryPortalMeterResponse(rsp)
}

// CreatePortalTokenWithBodyWithResponse request with arbitrary body returning *CreatePortalTokenResponse
func (c *ClientWithResponses) CreatePortalTokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreatePortalTokenResponse, error) {
	rsp, err := c.CreatePortalTokenWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreatePortalTokenResponse(rsp)
}

func (c *ClientWithResponses) CreatePortalTokenWithResponse(ctx context.Context, body CreatePortalTokenJSONRequestBody, reqEditors ...RequestEditorFn) (*CreatePortalTokenResponse, error) {
	rsp, err := c.CreatePortalToken(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreatePortalTokenResponse(rsp)
}

// InvalidatePortalTokensWithBodyWithResponse request with arbitrary body returning *InvalidatePortalTokensResponse
func (c *ClientWithResponses) InvalidatePortalTokensWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*InvalidatePortalTokensResponse, error) {
	rsp, err := c.InvalidatePortalTokensWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseInvalidatePortalTokensResponse(rsp)
}

func (c *ClientWithResponses) InvalidatePortalTokensWithResponse(ctx context.Context, body InvalidatePortalTokensJSONRequestBody, reqEditors ...RequestEditorFn) (*InvalidatePortalTokensResponse, error) {
	rsp, err := c.InvalidatePortalTokens(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseInvalidatePortalTokensResponse(rsp)
}

// ParseListEventsResponse parses an HTTP response from a ListEventsWithResponse call
func ParseListEventsResponse(rsp *http.Response) (*ListEventsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListEventsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []IngestedEvent
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseIngestEventsResponse parses an HTTP response from a IngestEventsWithResponse call
func ParseIngestEventsResponse(rsp *http.Response) (*IngestEventsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &IngestEventsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseListMetersResponse parses an HTTP response from a ListMetersWithResponse call
func ParseListMetersResponse(rsp *http.Response) (*ListMetersResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListMetersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Meter
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseCreateMeterResponse parses an HTTP response from a CreateMeterWithResponse call
func ParseCreateMeterResponse(rsp *http.Response) (*CreateMeterResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateMeterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest Meter
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 501:
		var dest NotImplementedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON501 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseDeleteMeterResponse parses an HTTP response from a DeleteMeterWithResponse call
func ParseDeleteMeterResponse(rsp *http.Response) (*DeleteMeterResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteMeterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest NotFoundProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 501:
		var dest NotImplementedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON501 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseGetMeterResponse parses an HTTP response from a GetMeterWithResponse call
func ParseGetMeterResponse(rsp *http.Response) (*GetMeterResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetMeterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Meter
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest NotFoundProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseQueryMeterResponse parses an HTTP response from a QueryMeterWithResponse call
func ParseQueryMeterResponse(rsp *http.Response) (*QueryMeterResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &QueryMeterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Data []MeterQueryRow `json:"data"`
			From *time.Time      `json:"from,omitempty"`
			To   *time.Time      `json:"to,omitempty"`

			// WindowSize Aggregation window size.
			WindowSize *WindowSize `json:"windowSize,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	case rsp.StatusCode == 200:
		// Content-type (text/csv) unsupported

	}

	return response, nil
}

// ParseListMeterSubjectsResponse parses an HTTP response from a ListMeterSubjectsWithResponse call
func ParseListMeterSubjectsResponse(rsp *http.Response) (*ListMeterSubjectsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListMeterSubjectsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseQueryPortalMeterResponse parses an HTTP response from a QueryPortalMeterWithResponse call
func ParseQueryPortalMeterResponse(rsp *http.Response) (*QueryPortalMeterResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &QueryPortalMeterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Data []MeterQueryRow `json:"data"`
			From *time.Time      `json:"from,omitempty"`
			To   *time.Time      `json:"to,omitempty"`

			// WindowSize Aggregation window size.
			WindowSize *WindowSize `json:"windowSize,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest UnauthorizedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	case rsp.StatusCode == 200:
		// Content-type (text/csv) unsupported

	}

	return response, nil
}

// ParseCreatePortalTokenResponse parses an HTTP response from a CreatePortalTokenWithResponse call
func ParseCreatePortalTokenResponse(rsp *http.Response) (*CreatePortalTokenResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreatePortalTokenResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest PortalToken
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseInvalidatePortalTokensResponse parses an HTTP response from a InvalidatePortalTokensWithResponse call
func ParseInvalidatePortalTokensResponse(rsp *http.Response) (*InvalidatePortalTokensResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &InvalidatePortalTokensResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest BadRequestProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest UnexpectedProblemResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{
	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
