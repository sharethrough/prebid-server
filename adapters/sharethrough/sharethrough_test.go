package sharethrough

import (
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"net/http"
	"testing"
)

type SharethroughAdapterInterface interface {
	generateHBUri(string, hbUriParams, *openrtb.App) string
}

type MockSharethroughAdapter struct {
	SharethroughAdapterInterface
	*SharethroughAdapter
}

func (s MockSharethroughAdapter) generateHBUri(baseUrl string, params hbUriParams, app *openrtb.App) string {
	return "http://ppp.com?placement_key=pkey"
}

//func (s MockSharethroughAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
//	return s.adapter.MakeRequests(request)
//}

func assertRequestDataEquals(t *testing.T, testName string, expected []*adapters.RequestData, actual []*adapters.RequestData) {
	t.Logf("Test case: %s\n", testName)

	if len(expected) != len(actual) {
		t.Errorf("Expected %d requests, got %d\n", len(expected), len(actual))
	}

	for index, expectedReq := range expected {
		if expectedReq.Method != actual[index].Method {
			t.Errorf("Method mismatch: expected %s got %s\n", expectedReq.Method, actual[index].Method)
		}
		if expectedReq.Uri != actual[index].Uri {
			t.Errorf("Uri mismatch: expected %s got %s\n", expectedReq.Uri, actual[index].Uri)
		}
		if len(expectedReq.Body) != len(actual[index].Body) {
			t.Errorf("Body mismatch: expected %s got %s\n", expectedReq.Body, actual[index].Body)
		}
		for headerIndex, expectedHeader := range expectedReq.Headers {
			if expectedHeader[0] != actual[index].Headers[headerIndex][0] {
				t.Errorf("Header %s mismatch: expected %s got %s\n", headerIndex, expectedHeader[0], actual[index].Headers[headerIndex][0])
			}
		}
	}

}

func TestSuccessMakeRequests(t *testing.T) {
	tests := map[string]struct {
		input    *openrtb.BidRequest
		expected []*adapters.RequestData
	}{
		"Generates expected Request": {
			input: &openrtb.BidRequest{
				App: &openrtb.App{Ext: []byte(`{}`)},
				Device: &openrtb.Device{
					UA: "Android Chome/60",
				},
				Imp: []openrtb.Imp{{
					ID:  "abc",
					Ext: []byte(`{"pkey": "pkey", "iframe": true, "iframeSize": [10, 20]}`),
					Banner: &openrtb.Banner{
						Format: []openrtb.Format{{H: 30, W: 40}},
					},
				}},
			},
			expected: []*adapters.RequestData{{
				Method: "POST",
				Uri:    "http://abc.com?placement_key=pkey&bidId=bid&consent_required=true&consent_string=consent&instant_play_capable=true&stayInIframe=false&height=20&width=30&hbVersion=1&supplyId=FGMrCMMc&strVersion=1.0.0",
				Body:   nil,
				Headers: http.Header{
					"Content-Type": []string{"text/plain;charset=utf-8"},
					"Accept":       []string{"application/json"},
				},
			}},
		},
	}

	for testName, test := range tests {
		t.Logf("Test case: %s\n", testName)

		adapter := &MockSharethroughAdapter{nil, &SharethroughAdapter{URI: "http://abc.com"}}
		output, actualErrors := adapter.MakeRequests(test.input)

		assertRequestDataEquals(t, testName, test.expected, output)
		if len(actualErrors) != 0 {
			t.Errorf("Expected no errors, got %d\n", len(actualErrors))
		}
	}

}
