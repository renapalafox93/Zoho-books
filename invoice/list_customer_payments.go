package invoice

import (
	"fmt"
	"github.com/schmorrison/Zoho"
)

//https://www.zoho.com/invoice/api/v3/#Recurring_Invoices_List_Recurring_Invoice
//func (c *ZohoInvoiceAPI) ListCustomerPayments(request interface{}, organizationId string, params map[string]zoho.Parameter) (data ListCustomerPaymentsResponse, err error) {
func (c *ZohoInvoiceAPI) ListCustomerPayments() (data ListCustomerPaymentsResponse, err error) {

	// Renew token if necessary
	if c.Zoho.Token.CheckExpiry() {
		err := c.Zoho.RefreshTokenRequest()
		if err != nil {
			return ListCustomerPaymentsResponse{}, err
		}
	}

	endpoint := zoho.Endpoint{
		Name:         CustomerPaymentsModule,
		URL:          fmt.Sprintf(zoho.InvoiceAPIEndPoint + "%s", CustomerPaymentsModule),
		Method:       zoho.HTTPGet,
		ResponseData: &ListCustomerPaymentsResponse{},
		URLParameters: map[string]zoho.Parameter{
			//"filter_by": "",
		},
	}

	/*for k, v := range params {
		endpoint.URLParameters[k] = v
	}
	*/

	err = c.Zoho.HTTPRequest(&endpoint)
	if err != nil {
		return ListCustomerPaymentsResponse{}, fmt.Errorf("Failed to retrieve expense reports: %s", err)
	}

	if v, ok := endpoint.ResponseData.(*ListCustomerPaymentsResponse); ok {
		// Check if the request succeeded
		if v.Code != 0 {
			return *v, fmt.Errorf("Failed to list customer payments: %s", v.Message)
		}
		return *v, nil
	}
	return ListCustomerPaymentsResponse{}, fmt.Errorf("Data retrieved was not 'ListCustomerPaymentsResponse'")
}

type ListCustomerPaymentsResponse struct {
	Code             int    `json:"code"`
	Message          string `json:"message"`
	CustomerPayments []struct {
		PaymentId     string  `json:"payment_id"`
		PaymentNumber string  `json:"payment_number"`
		InvoiceNumber string  `json:"invoice_number"`
		Date          string  `json:"date"`
		PaymentMode   string  `json:"payment_mode"`
		Amount        float64 `json:"amount"`
		BcyAmount     float64 `json:"bcy_amount"`
	} `json:"customerpayments"`
}
