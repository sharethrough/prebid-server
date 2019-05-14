package sharethrough

import (
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/url"
	"strconv"
)

type hbUriParams struct {
	Pkey               string
	BidID              string
	ConsentRequired    bool
	ConsentString      string
	InstantPlayCapable bool
	Iframe             bool
	Height             uint64
	Width              uint64
}

func butlerToOpenRTBResponse(btlrReq *adapters.RequestData, strResp openrtb_ext.ExtImpSharethroughResponse) (*adapters.BidderResponse, []error) {
	var errs []error
	bidResponse := adapters.NewBidderResponse()

	bidResponse.Currency = "USD"
	typedBid := &adapters.TypedBid{BidType: openrtb_ext.BidTypeNative}

	if len(strResp.Creatives) == 0 {
		errs = append(errs, &errortypes.BadInput{Message: "No creative provided"})
		return nil, errs
	}
	creative := strResp.Creatives[0]

	btlrParams, parseHBUriErr := parseHBUri(btlrReq.Uri)
	if parseHBUriErr != nil {
		errs = append(errs, &errortypes.BadInput{Message: parseHBUriErr.Error()})
		return nil, errs
	}

	util := &Util{}
	adm, admErr := util.getAdMarkup(strResp, btlrParams)
	if admErr != nil {
		errs = append(errs, &errortypes.BadServerResponse{Message: admErr.Error()})
	}

	bid := &openrtb.Bid{
		AdID:   strResp.AdServerRequestID,
		ID:     strResp.BidID,
		ImpID:  btlrParams.BidID,
		Price:  creative.CPM,
		CID:    creative.Metadata.CampaignKey,
		CrID:   creative.Metadata.CreativeKey,
		DealID: creative.Metadata.DealID,
		AdM:    adm,
		H:      btlrParams.Height,
		W:      btlrParams.Width,
	}

	typedBid.Bid = bid
	bidResponse.Bids = append(bidResponse.Bids, typedBid)

	return bidResponse, errs
}

func parseHBUri(uri string) (*hbUriParams, error) {
	btlrUrl, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	params := btlrUrl.Query()
	height, err := strconv.ParseUint(params.Get("height"), 10, 64)
	if err != nil {
		return nil, err
	}

	width, err := strconv.ParseUint(params.Get("width"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &hbUriParams{
		Pkey:            params.Get("placement_key"),
		BidID:           params.Get("bidId"),
		Iframe:          params.Get("stayInIframe") == "true",
		Height:          height,
		Width:           width,
		ConsentRequired: params.Get("consent_required") == "true",
		ConsentString:   params.Get("consent_string"),
	}, nil
}
