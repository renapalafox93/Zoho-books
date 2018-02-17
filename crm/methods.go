package crm

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"time"

	"github.com/schmorrison/Zoho"
)

const (
	baseURL = "https://crm.zoho.com/crm/private/"
	xmlURL  = baseURL + "xml/"
	jsonURL = baseURL + "json/"
)

//API is used for interacting with the Zoho CRM API
// the exposed methods are primarily access to CRM modules which provide access to CRM Methods
type API struct {
	*zoho.Zoho
	id string
}

//New returns a *crm.API with the provided zoho.Zoho as an embedded field
func New(z *zoho.Zoho) *API {
	id := func() string {
		var id []byte
		keyspace := "abcdefghijklmnopqrutuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for i := 0; i < 25; i++ {
			id = append(id, keyspace[rand.Intn(len(keyspace))])
		}
		return string(id)
	}()
	return &API{
		Zoho: z,
		id:   id,
	}
}

// makeRequest will perform the HTTP(s) request to Zoho CRM API with the provided authtoken and scope
// 
func (a *API) makeRequest(module crmModule, resource, method string, options optionEncoder) (crmData, error) {
	//make the URL and parse
	u := xmlURL + string(module) + resource
	U, err := url.Parse(string(u))
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL for method '%s' in module '%s': %s", resource, string(module), err.Error())
	}
	//encode the options to the URL
	options.encodeURL(U)

	U.RawQuery += "&authtoken=" + a.Zoho.GetAuthToken()
	
	//use 'zoho' module to make request/authenticate
	zr := a.Zoho.NewRequest(U.String(), "GET")

	if t := a.Zoho.GetAuthToken(); t != "" {
		zr.Add("authtoken", t)
	}

	resp, err := a.Zoho.Request(zr)
	if err != nil {
		return nil, fmt.Errorf("Error performing request for method '%s' in module '%s': %s", resource, string(module), err.Error())
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response to request for method '%s' in module '%s': %s", resource, string(module), err.Error())
	}

	//Get the 'type'
	data := module.getType()

	//decode response into CrmData
	data, err = decodeXML(b, data)

	//check if data returned is 'error' type
	if v, ok := data.(CrmError); ok {
		return nil, fmt.Errorf("Zoho returned an response of type '%s': %d: %s", v.Type, v.Code, v.Message)
	}

	return data, err
}

//Each module in Zoho CRM is a CrmModule
type CrmModule struct {
	api     *API
	id      string
	options optionEncoder
	module  crmModule
}

//Each modules CrmModule should provide the following methods
type crmMethoder interface {
	GetMyRecords(GetRecordsOptions) (crmData, error)
	GetRecords(GetRecordsOptions) (crmData, error)
	GetRecordById(GetRecordsByIdOptions) (crmData, error)
	GetDeletedRecordIds(GetDeletedRecordIdsOptions) (crmData, error)
	InsertRecords(InsertRecordsOptions) (crmData, error)
	UpdateRecords(UpdateRecordsOptions) (crmData, error)
	DeleteRecords(string) (crmData, error)
	GetSearchRecordsByPDC(GetSearchRecordsByPDCOptions) (crmData, error)
	ConvertLead(string, crmData) (crmData, error)
	GetRelatedRecords(GetRelatedRecordsOptions) (crmData, error)
	UpdateRelatedRecord(UpdateRelatedRecordOptions) (crmData, error)
	GetFields(int) (crmData, error)
	GetUsers(string) (crmData, error)
	UploadFile(UploadFileOptions) (crmData, error)
	DownloadFile(string) (crmData, error)
	DeleteFile(string) (crmData, error)
	Delink(DelinkOptions) (crmData, error)
	UploadPhoto(UploadPhotoOptions) (crmData, error)
	DownloadPhoto(string) (crmData, error)
	DeletePhoto(string) (crmData, error)
	GetModules(string) (crmData, error)
	SearchRecords(SearchRecordsOptions) (crmData, error)
}

// getMyRecords will make a request to the CrmModule specified using the options
// specified and return the type defined by the CrmModule
// https://www.zoho.com/crm/help/api/getmyrecords.html
func (a *API) getMyRecords(module crmModule, options GetRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/getMyRecords", "GET", options)
}

// getRecords will make a request to the CrmModule specified using the options
// specified and return the type defined by the CrmModule
// https://www.zoho.com/crm/help/api/getrecords.html
func (a *API) getRecords(module crmModule, options GetRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/getRecords", "GET", options)
}

// The options allowed for 'GetRecords' method
type GetRecordsOptions struct {
	SelectColumns    string    `zoho:"selectColumns,default>All"`   // To select the required fields from CRM module.
	FromIndex        int       `zoho:"fromIndex,default>1"`         // Default 1
	ToIndex          int       `zoho:"toIndex,default>20"`          // Default 20 // Maximum range value - 200
	SortColumnString string    `zoho:"sortColumnString"`            // You can select one of the fields in CRM in to sort the data.
	SortOrderString  string    `zoho:"sortOrderString,default>asc"` // Sorting order: asc or desc
	LastModifiedTime time.Time `zoho:"lastModifiedTime"`            // yyyy-MM-dd HH:mm:ss
}

func (o GetRecordsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// getRecordById will make a request to the CrmModule specified using the options
// specified and return the type defined by the CrmModule
// https://www.zoho.com/crm/help/api/getrecordbyid.html
func (a *API) getRecordById(module crmModule, options GetRecordsByIdOptions) (crmData, error) {
	return a.makeRequest(module, "/getRecordById", "GET", options)
}

// The options allowed for 'GetRecordsByID' method
type GetRecordsByIdOptions struct {
	ID     string   `zoho:"id"`
	IDList []string `zoho:"idlist"` // {id1};{id2} You can specify up to 100 IDs using this parameter.
}

func (o GetRecordsByIdOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// getDeletedRecordsIds is used to get the previously deleted records in a module
// https://www.zoho.com/crm/help/api/getdeletedrecordids.html
func (a *API) getDeletedRecordIds(module crmModule, options GetDeletedRecordIdsOptions) (crmData, error) {
	return a.makeRequest(module, "/getDeletedRecordIds", "GET", options)
}

type GetDeletedRecordIdsOptions struct {
	FromIndex        int       `zoho:"fromIndex,default>1"` // Default 1
	ToIndex          int       `zoho:"toIndex,default>20"`  // Default 20 // Maximum range value - 200
	LastModifiedTime time.Time `zoho:"lastModifiedTime"`    // yyyy-MM-dd HH:mm:ss
}

func (o GetDeletedRecordIdsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// insertRecords is used to create a record in a module
// https://www.zoho.com/crm/help/api/insertrecords.html
func (a *API) insertRecords(module crmModule, options InsertRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/insertRecords", "POST", options)
}

type DuplicateCheck int

const (
	ErrorOnDuplicate DuplicateCheck = 1 + iota
	UpdateOnDuplicate
)

type InsertRecordsOptions struct {
	Data            crmData        `zoho:"xmlData,required"`         // Required
	WorkflowTrigger bool           `zoho:"wfTrigger,default>false"`  // Set true to trigger workflow associated with the module
	DuplicateCheck  DuplicateCheck `zoho:"duplicateCheck,default>1"` // True will update the duplicate record, false will return an error
	IsApproval      bool           `zoho:"isApproval,default>false"` // True will require approval in CRM, available in Leads, Contacts, Cases
}

func (o InsertRecordsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// updateRecords is used to update a record in a module by ID
// https://www.zoho.com/crm/help/api/updaterecords.html
func (a *API) updateRecords(module crmModule, options UpdateRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/updateRecords", "POST", options)
}

type UpdateRecordsOptions struct {
	ID              string  `zoho:"id,required"`             // Required
	Data            crmData `zoho:"xmlData,required"`        // Required
	WorkflowTrigger bool    `zoho:"wfTrigger,default>false"` // Set true to trigger workflow associated with the module
}

func (o UpdateRecordsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// deleteRecords is used to delete a record from a module by ID
// https://www.zoho.com/crm/help/api/deleterecords.html
func (a *API) deleteRecords(module crmModule, id string) (crmData, error) {
	return a.makeRequest(module, "/deleteRecords", "GET", blankOptions{"id": id})
}

// getSearchByPDC is used to search records in a module and return only the 'pre-defined columns'
// https://www.zoho.com/crm/help/api/getsearchrecordsbypdc.html
func (a *API) getSearchRecordsByPDC(module crmModule, options GetSearchRecordsByPDCOptions) (crmData, error) {
	return a.makeRequest(module, "/getSearchRecordsByPDC", "GET", options)
}

type GetSearchRecordsByPDCOptions struct {
	SelectColumns string `zoho:"selectColumns,default>All"` // To select the required fields from CRM module.
	FromIndex     int    `zoho:"fromIndex,default>1"`       // Default 1
	ToIndex       int    `zoho:"toIndex,default>20"`        // Default 20 // Maximum range value - 200
	SearchColumn  string `zoho:"searchColumn,required,noencode"`     // Specify the predefined search column
	SearchValue   string `zoho:"searchValue,required,noencode"`      // Specify the value to be searched
}

func (o GetSearchRecordsByPDCOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// convertLead is used to change a Lead to a Contact/Potential/Account and remove the Lead
// https://www.zoho.com/crm/help/api/convertlead.html
func (a *API) convertLead(leadID string, data crmData) (crmData, error) {
	return a.makeRequest(leadsModule, "/convertLead", "GET", blankOptions{"leadId": leadID, "xmlData": fmt.Sprint(data)})
}

// getRelatedRecords is used to return the records that have been associated with the other record
// https://www.zoho.com/crm/help/api/getrelatedrecords.html
func (a *API) getRelatedRecords(module crmModule, options GetRelatedRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/getRelatedRecords", "GET", options)
}

type GetRelatedRecordsOptions struct {
	ParentModule crmModule `zoho:"parentModule,required"` // to fetch Leads related to a Campaign, then Campaigns is your parent module.
	ID           string    `zoho:"id,required"`           // id of the record for which you want to fetch related records
	FromIndex    int       `zoho:"fromIndex,default>1"`   // Default 1
	ToIndex      int       `zoho:"toIndex,default>20"`    // Default 20 // Maximum range value - 200
}

func (o GetRelatedRecordsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// updateRelatedRecord is used to update a record associated with a different record
// https://www.zoho.com/crm/help/api/updaterelatedrecords.html
func (a *API) updateRelatedRecord(module crmModule, options UpdateRelatedRecordOptions) (crmData, error) {
	return a.makeRequest(module, "/updateRelatedRecords", "GET", options)
}

type UpdateRelatedRecordOptions struct {
	ID            string    `zoho:"id,required"`
	RelatedModule crmModule `zoho:"relatedModule,required"`
	Data          crmData   `zoho:"xmlData,required"`
}

func (o UpdateRelatedRecordOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// getFields is used to retreive available fields from a CRM module
// kind: 1 - retrieve all fields , 2 - retrieve mandatory fields
// https://www.zoho.com/crm/help/api/getfields.html
func (a *API) getFields(module crmModule, kind int) (crmData, error) {
	return a.makeRequest(module, "/getFields", "GET", blankOptions{"type": fmt.Sprintf("%d", kind)})
}

// getUsers is used to return CRM users with provided status
// kind: AllUsers, ActiveUsers, DeactiveUsers, AdminUsers, ActiveConfirmedAdmins
// https://www.zoho.com/crm/help/api/getusers.html
func (a *API) getUsers(kind string) (crmData, error) {
	return a.makeRequest("Users", "/getUsers", "GET", blankOptions{"type": kind})
}

// uploadFile is used to associate a file with a record
// https://www.zoho.com/crm/help/api/uploadfile.html
func (a *API) uploadFile(module crmModule, options UploadFileOptions) (crmData, error) {
	return a.makeRequest(module, "/uploadFile", "GET", options)
}

type UploadFileOptions struct {
	ID            string `zoho:"id,required"`   // Required
	Content       string `zoho:"content"`       // File path
	AttachmentURL string `zoho:"attachmentUrl"` //URL of file
}

func (o UploadFileOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// downloadFile is used to download a File associated with a record
// https://www.zoho.com/crm/help/api/downloadfile.html
func (a *API) downloadFile(module crmModule, id string) (crmData, error) {
	return a.makeRequest(module, "/downloadFile", "GET", blankOptions{"id": id})
}

// deleteFile is used to delete a File associated with a record
// https://www.zoho.com/crm/help/api/deletefile.html
func (a *API) deleteFile(module crmModule, id string) (crmData, error) {
	return a.makeRequest(module, "/deleteFile", "GET", blankOptions{"id": id})
}

// delink is used to 'unassociate' two records
// https://www.zoho.com/crm/help/api/delink.html
func (a *API) delink(module crmModule, options DelinkOptions) (crmData, error) {
	return a.makeRequest(module, "/delink", "GET", options)
}

type DelinkOptions struct {
	ID            string    `zoho:"id,required"`            // Specify unique ID of the record.
	RelatedID     string    `zoho:"relatedId,required"`     // Specify unique ID of the child record.
	RelatedModule crmModule `zoho:"relatedModule,required"` // Specify name of the related module.
}

func (o DelinkOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// uploadPhoto is used to associate a photo with a record
// https://www.zoho.com/crm/help/api/uploadphoto.html
func (a *API) uploadPhoto(module crmModule, options UploadPhotoOptions) (crmData, error) {
	return a.makeRequest(module, "/uploadPhoto", "GET", options)
}

type UploadPhotoOptions struct {
	ID      string `zoho:"id,required"`      // Specify unique ID of the record.
	Content string `zoho:"content,required"` // File location
}

func (o UploadPhotoOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}

// downloadPhoto is used to download the photo associated with a record
// https://www.zoho.com/crm/help/api/downloadphoto.html
func (a *API) downloadPhoto(module crmModule, id string) (crmData, error) {
	return a.makeRequest(module, "/downloadPhoto", "GET", blankOptions{"id": id})
}

// deletePhoto is used to remove a photo associated with a record
// https://www.zoho.com/crm/help/api/deletephoto.html
func (a *API) deletePhoto(module crmModule, id string) (crmData, error) {
	return a.makeRequest(module, "/deletePhoto", "GET", blankOptions{"id": id})
}

//specify kind: 'api' will return API accessible modules
// https://www.zoho.com/crm/help/api/getmodules.html
func (a *API) getModules(kind string) (crmData, error) {
	return a.makeRequest("Info", "/getModules", "GET", blankOptions{"type": kind})
}

// searchRecords will return the records found during search
// https://www.zoho.com/crm/help/api/searchrecords.html
func (a *API) searchRecords(module crmModule, options SearchRecordsOptions) (crmData, error) {
	return a.makeRequest(module, "/searchRecords", "GET", options)
}

type SearchRecordsOptions struct {
	Criteria         string    `zoho:"criteria,required,noencode"`         // Specify the search criteria as shown in the Request URL section.
	SelectColumns    string    `zoho:"selectColumns,default>All"` //
	FromIndex        int       `zoho:"fromIndex,default>1"`       // Default 1
	ToIndex          int       `zoho:"toIndex,default>20"`        // Default 20 // Maximum range value - 200
	LastModifiedTime time.Time `zoho:"lastModifiedTime"`          // yyyy-MM-dd HH:mm:ss
}

func (o SearchRecordsOptions) encodeURL(u *url.URL) error {
	return encodeOptionsToURL(o, u)
}
