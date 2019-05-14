package sharethrough

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/openrtb_ext"
	"html/template"
	"net/url"
	"regexp"
	"strconv"
)

const minChromeVersion = 53
const minSafariVersion = 10

type UtilIface interface {
	canAutoPlayVideo(string) bool
	gdprApplies(*openrtb.BidRequest) bool
	gdprConsentString(*openrtb.BidRequest) string
	generateHBUri(string, hbUriParams, *openrtb.App) string
	getAdMarkup(openrtb_ext.ExtImpSharethroughResponse, *hbUriParams) (string, error)
	getPlacementSize([]openrtb.Format) (uint64, uint64)
	isAndroid(string) bool
	isAtMinChromeIosVersion(string) bool
	isAtMinChromeVersion(string) bool
	isAtMinSafariVersion(string) bool
	isiOS(string) bool
}

type Util struct{}

func (u Util) generateHBUri(baseUrl string, params hbUriParams, app *openrtb.App) string {
	v := url.Values{}
	v.Set("placement_key", params.Pkey)
	v.Set("bidId", params.BidID)
	v.Set("consent_required", fmt.Sprintf("%t", params.ConsentRequired))
	v.Set("consent_string", params.ConsentString)

	v.Set("instant_play_capable", fmt.Sprintf("%t", params.InstantPlayCapable))
	v.Set("stayInIframe", fmt.Sprintf("%t", params.Iframe))
	v.Set("height", strconv.FormatUint(params.Height, 10))
	v.Set("width", strconv.FormatUint(params.Width, 10))

	var version string

	if app != nil {
		// Skipping error handling here because it should fall through to unknown in the flow
		version, _ = jsonparser.GetString(app.Ext, "prebid", "version")
	}

	if len(version) == 0 {
		version = "unknown"
	}

	v.Set("hbVersion", version)
	v.Set("supplyId", supplyId)
	v.Set("strVersion", strVersion)

	return baseUrl + "?" + v.Encode()
}

func (u Util) getAdMarkup(strResp openrtb_ext.ExtImpSharethroughResponse, params *hbUriParams) (string, error) {
	strRespId := fmt.Sprintf("str_response_%s", strResp.BidID)
	jsonPayload, err := json.Marshal(strResp)
	if err != nil {
		return "", err
	}

	tmplBody := `
		<img src="//b.sharethrough.com/butler?type=s2s-win&arid={{.Arid}}" />

		<div data-str-native-key="{{.Pkey}}" data-stx-response-name="{{.StrRespId}}"></div>
	 	<script>var {{.StrRespId}} = "{{.B64EncodedJson}}"</script>
	`

	if params.Iframe {
		tmplBody = tmplBody + `
			<script src="//native.sharethrough.com/assets/sfp.js"></script>
		`
	} else {
		tmplBody = tmplBody + `
			<script src="//native.sharethrough.com/assets/sfp-set-targeting.js"></script>
	    	<script>
		     (function() {
		       if (!(window.STR && window.STR.Tag) && !(window.top.STR && window.top.STR.Tag)) {
		         var sfp_js = document.createElement('script');
		         sfp_js.src = "//native.sharethrough.com/assets/sfp.js";
		         sfp_js.type = 'text/javascript';
		         sfp_js.charset = 'utf-8';
		         try {
		             window.top.document.getElementsByTagName('body')[0].appendChild(sfp_js);
		         } catch (e) {
		           console.log(e);
		         }
		       }
		     })()
		   </script>
		`
	}

	tmpl, err := template.New("sfpjs").Parse(tmplBody)
	if err != nil {
		return "", err
	}

	var buf []byte
	templatedBuf := bytes.NewBuffer(buf)

	b64EncodedJson := base64.StdEncoding.EncodeToString(jsonPayload)
	err = tmpl.Execute(templatedBuf, struct {
		Arid           template.JS
		Pkey           string
		StrRespId      template.JS
		B64EncodedJson string
	}{
		template.JS(strResp.AdServerRequestID),
		params.Pkey,
		template.JS(strRespId),
		b64EncodedJson,
	})
	if err != nil {
		return "", err
	}

	return templatedBuf.String(), nil
}

func (u Util) getPlacementSize(formats []openrtb.Format) (height uint64, width uint64) {
	biggest := struct {
		Height uint64
		Width  uint64
	}{
		Height: 1,
		Width:  1,
	}

	for i := 0; i < len(formats); i++ {
		format := formats[i]
		if (format.H * format.W) > (biggest.Height * biggest.Width) {
			biggest.Height = format.H
			biggest.Width = format.W
		}
	}

	return biggest.Height, biggest.Width
}

func (u Util) canAutoPlayVideo(userAgent string) bool {
	return (u.isAndroid(userAgent) && u.isAtMinChromeVersion(userAgent)) ||
		(u.isiOS(userAgent) &&
			(u.isAtMinSafariVersion(userAgent) || u.isAtMinChromeIosVersion(userAgent))) ||
		!(u.isAndroid(userAgent) || u.isiOS(userAgent))
}

func (u Util) isAndroid(userAgent string) bool {
	isAndroid, err := regexp.MatchString("(?i)Android", userAgent)
	if err != nil {
		return false
	}
	return isAndroid
}

func (u Util) isiOS(userAgent string) bool {
	isiOS, err := regexp.MatchString("(?i)iPhone|iPad|iPod", userAgent)
	if err != nil {
		return false
	}
	return isiOS
}

func (u Util) isAtMinChromeVersion(userAgent string) bool {
	var chromeVersion int64
	var err error

	chromeVersionRegex := regexp.MustCompile(`Chrome\/(?P<ChromeVersion>\d+)`)
	chromeVersionMatch := chromeVersionRegex.FindStringSubmatch(userAgent)
	if len(chromeVersionMatch) > 1 {
		chromeVersion, err = strconv.ParseInt(chromeVersionMatch[1], 10, 64)
	}
	if err != nil {
		return false
	}

	return chromeVersion >= minChromeVersion
}

func (u Util) isAtMinChromeIosVersion(userAgent string) bool {
	var chromeiOSVersion int64
	var err error

	chromeiOSVersionRegex := regexp.MustCompile(`CriOS\/(?P<chromeiOSVersion>\d+)`)
	chromeiOSVersionMatch := chromeiOSVersionRegex.FindStringSubmatch(userAgent)
	if len(chromeiOSVersionMatch) > 1 {
		chromeiOSVersion, err = strconv.ParseInt(chromeiOSVersionMatch[1], 10, 64)
	}
	if err != nil {
		return false
	}

	return chromeiOSVersion >= minChromeVersion
}

func (u Util) isAtMinSafariVersion(userAgent string) bool {
	var safariVersion int64
	var err error

	safariVersionRegex := regexp.MustCompile(`Version\/(?P<safariVersion>\d+)`)
	safariVersionMatch := safariVersionRegex.FindStringSubmatch(userAgent)
	if len(safariVersionMatch) > 1 {
		safariVersion, err = strconv.ParseInt(safariVersionMatch[1], 10, 64)
	}
	if err != nil {
		return false
	}

	return safariVersion >= minSafariVersion
}

func (u Util) gdprApplies(request *openrtb.BidRequest) bool {
	var gdprApplies int64

	if request.Regs != nil {
		if jsonExtRegs, err := request.Regs.Ext.MarshalJSON(); err == nil {
			// 0 is the return value if error, so no need to handle
			gdprApplies, _ = jsonparser.GetInt(jsonExtRegs, "gdpr")
		}
	}

	return gdprApplies != 0
}

func (u Util) gdprConsentString(request *openrtb.BidRequest) string {
	var consentString string

	if request.User != nil {
		if jsonExtUser, err := request.User.Ext.MarshalJSON(); err == nil {
			// empty string is the return value if error, so no need to handle
			consentString, _ = jsonparser.GetString(jsonExtUser, "consent")
		}
	}

	return consentString
}
