package web

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator 0.12.0.0
// Changes may cause incorrect behavior and will be lost if the code is
// regenerated.

import (
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"net/http"
	"net/url"
)

// GlobalCertificateOrderClient is the use these APIs to manage Azure Websites
// resources through the Azure Resource Manager. All task operations conform
// to the HTTP/1.1 protocol specification and each operation returns an
// x-ms-request-id header that can be used to obtain information about the
// request. You must make sure that requests made to these resources are
// secure. For more information, see <a
// href="https://msdn.microsoft.com/en-us/library/azure/dn790557.aspx">Authenticating
// Azure Resource Manager requests.</a>
type GlobalCertificateOrderClient struct {
	SiteManagementClient
}

// NewGlobalCertificateOrderClient creates an instance of the
// GlobalCertificateOrderClient client.
func NewGlobalCertificateOrderClient(subscriptionID string) GlobalCertificateOrderClient {
	return NewGlobalCertificateOrderClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewGlobalCertificateOrderClientWithBaseURI creates an instance of the
// GlobalCertificateOrderClient client.
func NewGlobalCertificateOrderClientWithBaseURI(baseURI string, subscriptionID string) GlobalCertificateOrderClient {
	return GlobalCertificateOrderClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// GetAllCertificateOrders sends the get all certificate orders request.
func (client GlobalCertificateOrderClient) GetAllCertificateOrders() (result CertificateOrderCollection, ae error) {
	req, err := client.GetAllCertificateOrdersPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalCertificateOrderClient", "GetAllCertificateOrders", "Failure preparing request")
	}

	resp, err := client.GetAllCertificateOrdersSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalCertificateOrderClient", "GetAllCertificateOrders", "Failure sending request")
	}

	result, err = client.GetAllCertificateOrdersResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalCertificateOrderClient", "GetAllCertificateOrders", "Failure responding to request")
	}

	return
}

// GetAllCertificateOrdersPreparer prepares the GetAllCertificateOrders request.
func (client GlobalCertificateOrderClient) GetAllCertificateOrdersPreparer() (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.CertificateRegistration/certificateOrders"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetAllCertificateOrdersSender sends the GetAllCertificateOrders request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalCertificateOrderClient) GetAllCertificateOrdersSender(req *http.Request) (*http.Response, error) {
	return client.Send(req, http.StatusOK)
}

// GetAllCertificateOrdersResponder handles the response to the GetAllCertificateOrders request. The method always
// closes the http.Response Body.
func (client GlobalCertificateOrderClient) GetAllCertificateOrdersResponder(resp *http.Response) (result CertificateOrderCollection, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		autorest.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
