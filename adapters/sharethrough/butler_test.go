package sharethrough

import (
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"strings"
	"testing"
)

func assertBidderResponseEquals(t *testing.T, testName string, expected adapters.BidderResponse, actual adapters.BidderResponse) {
	t.Logf("Test case: %s\n", testName)
	if len(expected.Bids) != len(actual.Bids) {
		t.Errorf("Expected %d bids in BidResponse, got %d\n", len(expected.Bids), len(actual.Bids))
		return
	}
	for index, expectedTypedBid := range expected.Bids {
		if expectedTypedBid.BidType != actual.Bids[index].BidType {
			t.Errorf("Bid[%d]: Type mismatch, expected %s got %s\n", index, expectedTypedBid.BidType, actual.Bids[index].BidType)
		}
		if expectedTypedBid.Bid.AdID != actual.Bids[index].Bid.AdID {
			t.Errorf("Bid[%d]: AdID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.AdID, actual.Bids[index].Bid.AdID)
		}
		if expectedTypedBid.Bid.ID != actual.Bids[index].Bid.ID {
			t.Errorf("Bid[%d]: ID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.ID, actual.Bids[index].Bid.ID)
		}
		if expectedTypedBid.Bid.ImpID != actual.Bids[index].Bid.ImpID {
			t.Errorf("Bid[%d]: ImpID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.ImpID, actual.Bids[index].Bid.ImpID)
		}
		if expectedTypedBid.Bid.Price != actual.Bids[index].Bid.Price {
			t.Errorf("Bid[%d]: Price mismatch, expected %f got %f\n", index, expectedTypedBid.Bid.Price, actual.Bids[index].Bid.Price)
		}
		if expectedTypedBid.Bid.CID != actual.Bids[index].Bid.CID {
			t.Errorf("Bid[%d]: CID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.CID, actual.Bids[index].Bid.CID)
		}
		if expectedTypedBid.Bid.CrID != actual.Bids[index].Bid.CrID {
			t.Errorf("Bid[%d]: CrID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.CrID, actual.Bids[index].Bid.CrID)
		}
		if expectedTypedBid.Bid.DealID != actual.Bids[index].Bid.DealID {
			t.Errorf("Bid[%d]: DealID mismatch, expected %s got %s\n", index, expectedTypedBid.Bid.DealID, actual.Bids[index].Bid.DealID)
		}
		if expectedTypedBid.Bid.H != actual.Bids[index].Bid.H {
			t.Errorf("Bid[%d]: H mismatch, expected %d got %d\n", index, expectedTypedBid.Bid.H, actual.Bids[index].Bid.H)
		}
		if expectedTypedBid.Bid.W != actual.Bids[index].Bid.W {
			t.Errorf("Bid[%d]: W mismatch, expected %d got %d\n", index, expectedTypedBid.Bid.W, actual.Bids[index].Bid.W)
		}
	}
}

func TestSuccessButlerToOpenRTBResponse(t *testing.T) {
	tests := map[string]struct {
		inputButlerReq  *adapters.RequestData
		inputStrResp    openrtb_ext.ExtImpSharethroughResponse
		expectedSuccess *adapters.BidderResponse
		expectedErrors  []error
	}{
		"Generates expected openRTB bid response": {
			inputButlerReq: &adapters.RequestData{
				Uri: "http://uri.com?placement_key=pkey&bidId=bidid&height=20&width=30",
			},
			inputStrResp: openrtb_ext.ExtImpSharethroughResponse{
				AdServerRequestID: "arid",
				BidID:             "bid",
				Creatives: []openrtb_ext.ExtImpSharethroughCreative{{
					CPM: 10,
					Metadata: openrtb_ext.ExtImpSharethroughCreativeMetadata{
						CampaignKey: "cmpKey",
						CreativeKey: "creaKey",
						DealID:      "dealId",
					},
				}},
			},
			expectedSuccess: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{{
					BidType: openrtb_ext.BidTypeNative,
					Bid: &openrtb.Bid{
						AdID:   "arid",
						ID:     "bid",
						ImpID:  "bidid",
						Price:  10,
						CID:    "cmpKey",
						CrID:   "creaKey",
						DealID: "dealId",
						H:      20,
						W:      30,
					},
				}},
			},
			expectedErrors: []error{},
		},
	}

	for testName, test := range tests {
		outputSuccess, outputErrors := butlerToOpenRTBResponse(test.inputButlerReq, test.inputStrResp)
		assertBidderResponseEquals(t, testName, *test.expectedSuccess, *outputSuccess)
		if len(outputErrors) != len(test.expectedErrors) {
			t.Errorf("Expected %d errors, got %d\n", len(test.expectedErrors), len(outputErrors))
		}
	}
}

func TestFailButlerToOpenRTBResponse(t *testing.T) {
	tests := map[string]struct {
		inputButlerReq  *adapters.RequestData
		inputStrResp    openrtb_ext.ExtImpSharethroughResponse
		expectedSuccess *adapters.BidderResponse
		expectedErrors  []error
	}{
		"Returns nil if no creatives provided": {
			inputButlerReq: &adapters.RequestData{},
			inputStrResp: openrtb_ext.ExtImpSharethroughResponse{
				Creatives: []openrtb_ext.ExtImpSharethroughCreative{},
			},
			expectedSuccess: nil,
			expectedErrors: []error{
				&errortypes.BadInput{Message: "No creative provided"},
			},
		},
		"Returns nil if failed to parse Uri": {
			inputButlerReq: &adapters.RequestData{
				Uri: "wrong format url",
			},
			inputStrResp: openrtb_ext.ExtImpSharethroughResponse{
				Creatives: []openrtb_ext.ExtImpSharethroughCreative{{}},
			},
			expectedSuccess: nil,
			expectedErrors: []error{
				&errortypes.BadInput{Message: `strconv.ParseUint: parsing "": invalid syntax`},
			},
		},
	}

	for testName, test := range tests {
		t.Logf("Test case: %s\n", testName)
		outputSuccess, outputErrors := butlerToOpenRTBResponse(test.inputButlerReq, test.inputStrResp)

		if test.expectedSuccess != outputSuccess {
			t.Errorf("Expected result %+v, got %+v\n", test.expectedSuccess, outputSuccess)
		}

		if len(outputErrors) != len(test.expectedErrors) {
			t.Errorf("Expected %d errors, got %d\n", len(test.expectedErrors), len(outputErrors))
		}

		for index, expectedError := range test.expectedErrors {
			if fmt.Sprintf("%T", expectedError) != fmt.Sprintf("%T", outputErrors[index]) {
				t.Errorf("Error type mismatch, expected %T, got %T\n", expectedError, outputErrors[index])
			}
			if expectedError.Error() != outputErrors[index].Error() {
				t.Errorf("Expected error %s, got %s\n", expectedError.Error(), outputErrors[index].Error())
			}
		}
	}
}

func TestGenerateHBUri(t *testing.T) {
	tests := map[string]struct {
		inputUrl    string
		inputParams hbUriParams
		inputApp    *openrtb.App
		expected    []string
	}{
		"Generates expected URL, appending all params": {
			inputUrl: "http://abc.com",
			inputParams: hbUriParams{
				Pkey:               "pkey",
				BidID:              "bid",
				ConsentRequired:    true,
				ConsentString:      "consent",
				InstantPlayCapable: true,
				Iframe:             false,
				Height:             20,
				Width:              30,
			},
			inputApp: &openrtb.App{Ext: []byte(`{"prebid": {"version": "1"}}`)},
			expected: []string{
				"http://abc.com?",
				"placement_key=pkey",
				"bidId=bid",
				"consent_required=true",
				"consent_string=consent",
				"instant_play_capable=true",
				"stayInIframe=false",
				"height=20",
				"width=30",
				"hbVersion=1",
				"supplyId=FGMrCMMc",
				"strVersion=1.0.0",
			},
		},
		"Sets version to unknown if version not found": {
			inputUrl:    "http://abc.com",
			inputParams: hbUriParams{},
			inputApp:    &openrtb.App{Ext: []byte(`{}`)},
			expected: []string{
				"hbVersion=unknown",
			},
		},
	}

	adapter := NewSharethroughBidder("http://abc.com")

	for testName, test := range tests {
		t.Logf("Test case: %s\n", testName)
		output := adapter.generateHBUri(test.inputUrl, test.inputParams, test.inputApp)

		for _, uriParam := range test.expected {
			if !strings.Contains(output, uriParam) {
				t.Errorf("Expected %s to be found in URL, got %s\n", uriParam, output)
			}
		}
	}
}

func assertHbUriParamsEquals(t *testing.T, testName string, expected *hbUriParams, actual *hbUriParams) {
	t.Logf("Test case: %s\n", testName)
	if expected.Pkey != actual.Pkey {
		t.Errorf("Expected Pkey to be %s, got %s\n", expected.Pkey, actual.Pkey)
	}
	if expected.BidID != actual.BidID {
		t.Errorf("Expected BidID to be %s, got %s\n", expected.BidID, actual.BidID)
	}
	if expected.Iframe != actual.Iframe {
		t.Errorf("Expected Iframe to be %t, got %t\n", expected.Iframe, actual.Iframe)
	}
	if expected.Height != actual.Height {
		t.Errorf("Expected Height to be %d, got %d\n", expected.Height, actual.Height)
	}
	if expected.Width != actual.Width {
		t.Errorf("Expected Width to be %d, got %d\n", expected.Width, actual.Width)
	}
	if expected.ConsentRequired != actual.ConsentRequired {
		t.Errorf("Expected ConsentRequired to be %t, got %t\n", expected.ConsentRequired, actual.ConsentRequired)
	}
	if expected.ConsentString != actual.ConsentString {
		t.Errorf("Expected ConsentString to be %s, got %s\n", expected.ConsentString, actual.ConsentString)
	}
}

func TestSuccessParseHBUri(t *testing.T) {
	tests := map[string]struct {
		input           string
		expectedSuccess *hbUriParams
	}{
		"Decodes URI successfully": {
			input: "http://abc.com?placement_key=pkey&bidId=bid&consent_required=true&consent_string=consent&instant_play_capable=true&stayInIframe=false&height=20&width=30&hbVersion=1&supplyId=FGMrCMMc&strVersion=1.0.0",
			expectedSuccess: &hbUriParams{
				Pkey:            "pkey",
				BidID:           "bid",
				Iframe:          false,
				Height:          20,
				Width:           30,
				ConsentRequired: true,
				ConsentString:   "consent",
			},
		},
	}

	for testName, test := range tests {
		t.Logf("Test case: %s\n", testName)
		output, actualError := parseHBUri(test.input)

		assertHbUriParamsEquals(t, testName, test.expectedSuccess, output)
		if actualError != nil {
			t.Errorf("Expected no errors, got %s\n", actualError)
		}
	}
}

func TestFailParseHBUri(t *testing.T) {
	tests := map[string]struct {
		input         string
		expectedError string
	}{
		"Fails decoding if unable to parse URI": {
			input:         "wrong URI",
			expectedError: `strconv.ParseUint: parsing "": invalid syntax`,
		},
		"Fails decoding if height not provided": {
			input:         "http://abc.com?width=10",
			expectedError: `strconv.ParseUint: parsing "": invalid syntax`,
		},
		"Fails decoding if width not provided": {
			input:         "http://abc.com?height=10",
			expectedError: `strconv.ParseUint: parsing "": invalid syntax`,
		},
	}

	for testName, test := range tests {
		t.Logf("Test case: %s\n", testName)
		output, actualError := parseHBUri(test.input)

		if output != nil {
			t.Errorf("Expected return value nil, got %+v\n", output)
		}
		if actualError == nil {
			t.Errorf("Expected error not to be nil\n")
			break
		}
		if actualError.Error() != test.expectedError {
			t.Errorf("Expected error '%s', got '%s'\n", test.expectedError, actualError.Error())
		}
	}
}
