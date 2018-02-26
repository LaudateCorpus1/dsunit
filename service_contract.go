package dsunit

import (
	"errors"
	"fmt"
	"github.com/viant/assertly"
	"github.com/viant/dsc"
	"github.com/viant/toolbox/url"
)

//StatusOk represents ok status
const StatusOk = "ok"

//BaseResponse represent base response.
type BaseResponse struct {
	Status  string
	Message string
}

func (r BaseResponse) Error() error {
	if r.Status != StatusOk {
		return errors.New(r.Message)
	}
	return nil
}

func (r *BaseResponse) SetError(err error) {
	if err == nil {
		return
	}
	r.Status = "error"
	r.Message = err.Error()
}

func NewBaseResponse(status, message string) *BaseResponse {
	return &BaseResponse{
		Status:  status,
		Message: message,
	}
}

func NewBaseOkResponse() *BaseResponse {
	return NewBaseResponse(StatusOk, "")
}

//RegisterRequest represent register request
type RegisterRequest struct {
	Datastore string                 `required:"true" description:"datastore name"`
	Config    *dsc.Config            `description:"datastore config"`
	ConfigURL string                 `description:"datastore config URL"`
	Tables    []*dsc.TableDescriptor `description:"optional table descriptors"`
}

//NewRegisterRequest create new register request
func NewRegisterRequest(datastore string, config *dsc.Config, tables ...*dsc.TableDescriptor) *RegisterRequest {
	return &RegisterRequest{
		Datastore: datastore,
		Config:    config,
		Tables:    tables,
	}
}

func NewRegisterRequestFromURL(URL string) (*RegisterRequest, error) {
	var result = &RegisterRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//RegisterResponse represents register response
type RegisterResponse struct {
	*BaseResponse
}

//RecreateRequest represent recreate datastore request
type RecreateRequest struct {
	Datastore      string `required:"true" description:"datastore name to recreate, come database will create the whole schema, other will remove exiting tables and add registered one"`
	AdminDatastore string `description:"database  used to run DDL"`
}

//NewRecreateRequest create new recreate request
func NewRecreateRequest(datastore, adminDatastore string) *RecreateRequest {
	return &RecreateRequest{
		Datastore:      datastore,
		AdminDatastore: adminDatastore,
	}
}

//NewRecreateRequestFromURL create a request from URL
func NewRecreateRequestFromURL(URL string) (*RecreateRequest, error) {
	var result = &RecreateRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err

}

//RecreateResponse represents recreate datastore response
type RecreateResponse struct {
	*BaseResponse
}

//RunSQLRequest represents run SQL request
type RunSQLRequest struct {
	Datastore string `required:"true" description:"registered datastore name"`
	Expand    bool   `description:"substitute $ expression with content of context.state"`
	SQL       []string
}

//NewRunSQLRequest creates new run SQL request
func NewRunSQLRequest(datastore string, SQL ...string) *RunSQLRequest {
	return &RunSQLRequest{
		Datastore: datastore,
		SQL:       SQL,
	}
}

//NewRunSQLRequestFromURL create a request from URL
func NewRunSQLRequestFromURL(URL string) (*RunSQLRequest, error) {
	var result = &RunSQLRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//RunSQLRequest represents run SQL response
type RunSQLResponse struct {
	*BaseResponse
	RowsAffected int
}

//RunScriptRequest represents run SQL Script request
type RunScriptRequest struct {
	Datastore string `required:"true" description:"registered datastore name"`
	Expand    bool   `description:"substitute $ expression with content of context.state"`
	Scripts   []*url.Resource
}

//NewRunScriptRequest creates new run script request
func NewRunScriptRequest(datastore string, scripts ...*url.Resource) *RunScriptRequest {
	return &RunScriptRequest{
		Datastore: datastore,
		Scripts:   scripts,
	}
}

//NewRunScriptRequestFromURL create a request from URL
func NewRunScriptRequestFromURL(URL string) (*RunScriptRequest, error) {
	var result = &RunScriptRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//MappingRequest represnet a mapping request
type MappingRequest struct {
	Mappings []*Mapping `required:"true" description:"virtual table mapping"`
}

//Init init request
func (r *MappingRequest) Init() (err error) {
	if len(r.Mappings) == 0 {
		return nil
	}
	for _, mapping := range r.Mappings {
		if (mapping.Resource != nil && mapping.URL != "") || mapping.MappingTable == nil {
			if err = mapping.Init(); err == nil {
				err = mapping.JSONDecode(mapping)
			}
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (r *MappingRequest) Validate() error {
	if r == nil {
		return nil
	}
	if len(r.Mappings) == 0 {
		return errors.New("mappings were empty")
	}
	for i, mapping := range r.Mappings {
		if mapping.Name == "" {
			return fmt.Errorf("mappings[%v].name were empty", i)
		}
	}
	return nil
}

//NewMappingRequest creates new mapping request
func NewMappingRequest(mappings ...*Mapping) *MappingRequest {
	return &MappingRequest{
		Mappings: mappings,
	}
}

//NewMappingRequestFromURL create a request from URL
func NewMappingRequestFromURL(URL string) (*MappingRequest, error) {
	var result = &MappingRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//MappingResponse represents mapping response
type MappingResponse struct {
	*BaseResponse
	Tables []string
}

//InitRequest represents datastore init request, it actual aggregates, registraction, recreation, mapping and run script request
type InitRequest struct {
	Datastore string
	Recreate  bool
	*RegisterRequest
	Admin *RegisterRequest
	*MappingRequest
	*RunScriptRequest
}

func (r *InitRequest) Init() (err error) {
	if r.RegisterRequest != nil {
		if r.RegisterRequest.Datastore == "" {
			r.RegisterRequest.Datastore = r.Datastore
		}

		if r.RegisterRequest.Config == nil && r.RegisterRequest.ConfigURL != "" {
			r.Config, err = dsc.NewConfigFromURL(r.RegisterRequest.ConfigURL)
			if err != nil {
				return err
			}
		}
	}
	if r.RunScriptRequest != nil {
		if r.RunScriptRequest.Datastore == "" {
			r.RunScriptRequest.Datastore = r.Datastore
		}
	}

	return nil
}

func (r *InitRequest) Validate() error {
	if r.RegisterRequest == nil {
		return errors.New("datastore was empty")
	}
	if r.RegisterRequest.Config == nil {
		return errors.New("datastore config empty")
	}
	return nil
}

//NewInitRequest creates a new database init request
func NewInitRequest(datastore string, recreate bool, register, admin *RegisterRequest, mapping *MappingRequest, script *RunScriptRequest) *InitRequest {
	return &InitRequest{
		Datastore:        datastore,
		Recreate:         recreate,
		RegisterRequest:  register,
		Admin:            admin,
		MappingRequest:   mapping,
		RunScriptRequest: script,
	}
}

//NewInitRequestFromURL create a request from URL
func NewInitRequestFromURL(URL string) (*InitRequest, error) {
	var result = &InitRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//InitResponse represent init datastore response
type InitResponse struct {
	*BaseResponse
	Tables []string
}

//PrepareRequest represents a request to populate datastore with data resource
type PrepareRequest struct {
	Expand           bool `description:"substitute $ expression with content of context.state"`
	*DatasetResource `required:"true" description:"datasets resource"`
}

//Validate checks if request is valid
func (r *PrepareRequest) Validate() error {
	if r.Resource == nil {
		return errors.New("url was empty")
	}
	if r.DatastoreDatasets == nil {
		return errors.New("datastore was empty")
	}
	return nil
}

//NewPrepareRequest creates a new prepare request
func NewPrepareRequest(resource *DatasetResource) *PrepareRequest {
	return &PrepareRequest{
		DatasetResource: resource,
	}
}

//NewPrepareRequestFromURL create a request from URL
func NewPrepareRequestFromURL(URL string) (*PrepareRequest, error) {
	var result = &PrepareRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//ModificationInfo represents a modification info
type ModificationInfo struct {
	Subject  string
	Method   string `description:"modification method determined by presence of primary key: load - insert, persist: insert or update"`
	Deleted  int
	Modified int
	Added    int
}

//PrepareResponse represents a prepare response
type PrepareResponse struct {
	*BaseResponse
	Expand       bool                         `description:"substitute $ expression with content of context.state"`
	Modification map[string]*ModificationInfo `description:"modification info by subject"`
}

//ExpectRequest represents verification datastore request
type ExpectRequest struct {
	*DatasetResource
	CheckPolicy int `required:"true" description:"0 - FullTableDatasetCheckPolicy, 1 - SnapshotDatasetCheckPolicy"`
}

//Validate checks if request is valid
func (r *ExpectRequest) Validate() error {
	if r.Resource == nil {
		return errors.New("url was empty")
	}
	if r.DatastoreDatasets == nil {
		return errors.New("datastore was empty")
	}
	return nil
}

//NewExpectRequest creates a new prepare request
func NewExpectRequest(checkPolicy int, resource *DatasetResource) *ExpectRequest {
	return &ExpectRequest{
		CheckPolicy:     checkPolicy,
		DatasetResource: resource,
	}
}

//NewExpectRequestFromURL create a request from URL
func NewExpectRequestFromURL(URL string) (*ExpectRequest, error) {
	var result = &ExpectRequest{}
	resource := url.NewResource(URL)
	err := resource.JSONDecode(result)
	return result, err
}

//ExpectRequest represents data validation
type DatasetValidation struct {
	Dataset string
	*assertly.Validation
}

//ExpectResponse represents verification response
type ExpectResponse struct {
	*BaseResponse
	Validation  []*DatasetValidation
	PassedCount int
	FailedCount int
}

//SequenceRequest represents get sequences request
type SequenceRequest struct {
	Datastore string
	Tables    []string
}

func NewSequenceRequest(datastore string, tables ...string) *SequenceRequest {
	return &SequenceRequest{
		Datastore: datastore,
		Tables:    tables,
	}
}

//SequenceResponse represents get sequences response
type SequenceResponse struct {
	*BaseResponse
	Sequences map[string]int
}

//QueryRequest represents get sequences request
type QueryRequest struct {
	Datastore string
	SQL       string
}

func NewQueryRequest(datastore, SQL string) *QueryRequest {
	return &QueryRequest{
		Datastore: datastore,
		SQL:       SQL,
	}
}

//QueryResponse represents get sequences response
type QueryResponse struct {
	*BaseResponse
	Records Records
}
