package gdpr

import (
	"context"

	"github.com/prebid/go-gdpr/api"
	tcf1constants "github.com/prebid/go-gdpr/consentconstants"
	consentconstants "github.com/prebid/go-gdpr/consentconstants/tcf2"
	"github.com/prebid/go-gdpr/vendorconsent"
	"github.com/prebid/go-gdpr/vendorlist"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// This file implements GDPR permissions for the app.
// For more info, see https://github.com/prebid/prebid-server/issues/501
//
// Nothing in this file is exported. Public APIs can be found in gdpr.go

type permissionsImpl struct {
	cfg             config.GDPR
	vendorIDs       map[openrtb_ext.BidderName]uint16
	fetchVendorList map[uint8]func(ctx context.Context, id uint16) (vendorlist.VendorList, error)
}

func (p *permissionsImpl) HostCookiesAllowed(ctx context.Context, consent string) (bool, error) {
	return p.allowSync(ctx, uint16(p.cfg.HostVendorID), consent)
}

func (p *permissionsImpl) BidderSyncAllowed(ctx context.Context, bidder openrtb_ext.BidderName, consent string) (bool, error) {
	id, ok := p.vendorIDs[bidder]
	if ok {
		return p.allowSync(ctx, id, consent)
	}

	if consent == "" {
		return p.cfg.UsersyncIfAmbiguous, nil
	}

	return false, nil
}

func (p *permissionsImpl) PersonalInfoAllowed(ctx context.Context, bidder openrtb_ext.BidderName, PublisherID string, consent string) (bool, error) {
	_, ok := p.cfg.NonStandardPublisherMap[PublisherID]
	if ok {
		return true, nil
	}

	id, ok := p.vendorIDs[bidder]
	if ok {
		return p.allowPI(ctx, id, consent)
	}

	if consent == "" {
		return p.cfg.UsersyncIfAmbiguous, nil
	}

	return false, nil
}

func (p *permissionsImpl) allowSync(ctx context.Context, vendorID uint16, consent string) (bool, error) {
	// If we're not given a consent string, respect the preferences in the app config.
	if consent == "" {
		return p.cfg.UsersyncIfAmbiguous, nil
	}

	parsedConsent, vendor, err := p.parseVendor(ctx, vendorID, consent)
	if err != nil {
		return false, err
	}

	if vendor == nil {
		return false, nil
	}

	// InfoStorageAccess is the same across TCF 1 and TCF 2
	if vendor.Purpose(consentconstants.InfoStorageAccess) && parsedConsent.PurposeAllowed(consentconstants.InfoStorageAccess) && parsedConsent.VendorConsent(vendorID) {
		return true, nil
	}
	return false, nil
}

func (p *permissionsImpl) allowPI(ctx context.Context, vendorID uint16, consent string) (bool, error) {
	// If we're not given a consent string, respect the preferences in the app config.
	if consent == "" {
		return p.cfg.UsersyncIfAmbiguous, nil
	}

	parsedConsent, vendor, err := p.parseVendor(ctx, vendorID, consent)
	if err != nil {
		return false, err
	}

	if vendor == nil {
		return false, nil
	}

	if parsedConsent.Version() == 2 {
		// Need to add the location special purpose once the library supports it.
		if (vendor.Purpose(consentconstants.InfoStorageAccess) || vendor.LegitimateInterest(consentconstants.InfoStorageAccess)) && parsedConsent.PurposeAllowed(consentconstants.InfoStorageAccess) && (vendor.Purpose(consentconstants.PersonalizationProfile) || vendor.LegitimateInterest(consentconstants.PersonalizationProfile)) && parsedConsent.PurposeAllowed(consentconstants.PersonalizationProfile) && parsedConsent.VendorConsent(vendorID) {
			return true, nil
		}
	} else {
		if (vendor.Purpose(tcf1constants.InfoStorageAccess) || vendor.LegitimateInterest(tcf1constants.InfoStorageAccess)) && parsedConsent.PurposeAllowed(tcf1constants.InfoStorageAccess) && (vendor.Purpose(tcf1constants.AdSelectionDeliveryReporting) || vendor.LegitimateInterest(tcf1constants.AdSelectionDeliveryReporting)) && parsedConsent.PurposeAllowed(tcf1constants.AdSelectionDeliveryReporting) && parsedConsent.VendorConsent(vendorID) {
			return true, nil
		}
	}
	return false, nil
}

func (p *permissionsImpl) parseVendor(ctx context.Context, vendorID uint16, consent string) (parsedConsent api.VendorConsents, vendor api.Vendor, err error) {
	parsedConsent, err = vendorconsent.ParseString(consent)
	if err != nil {
		err = &ErrorMalformedConsent{
			consent: consent,
			cause:   err,
		}
		return
	}

	version := parsedConsent.Version()
	if version < 1 || version > 2 {
		return
	}
	vendorList, err := p.fetchVendorList[version](ctx, parsedConsent.VendorListVersion())
	if err != nil {
		return
	}

	vendor = vendorList.Vendor(vendorID)
	return
}

// Exporting to allow for easy test setups
type AlwaysAllow struct{}

func (a AlwaysAllow) HostCookiesAllowed(ctx context.Context, consent string) (bool, error) {
	return true, nil
}

func (a AlwaysAllow) BidderSyncAllowed(ctx context.Context, bidder openrtb_ext.BidderName, consent string) (bool, error) {
	return true, nil
}

func (a AlwaysAllow) PersonalInfoAllowed(ctx context.Context, bidder openrtb_ext.BidderName, PublisherID string, consent string) (bool, error) {
	return true, nil
}
