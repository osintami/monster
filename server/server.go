// © 2022 Sloan Childers
package server

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/osintami/monster/utils"
)

type MonsterServer struct {
	core    utils.ServerCore
	uaregex *regexp.Regexp
}

const (
	MY_COOKIE_ID     = "muid"
	ONE_YEAR_SECONDS = 60 * 60 * 24 * 365
)

var ErrEHashNotFound = errors.New("ehash not found")

func NewServer(core utils.ServerCore) *MonsterServer {
	return &MonsterServer{
		core:    core,
		uaregex: regexp.MustCompile(`useragent=([^&#]*)`)}
}

type CookieInfo struct {
	MyCookieID       string // found in cookies
	PartnerCookieID  string // query param
	PartnerID        string // query param
	PartnerEmailHash string // query param
	RedirectURL      string // query param
	UserAgent        string // found in header
	ClientIP         string // found in header
}

// Store the partner's user id (cookie id) and redirect to the endpoint of their choice with our cookie id.
func (x *MonsterServer) CookieSync(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().UnixMicro()
	var ci CookieInfo

	ci.PartnerCookieID = r.URL.Query().Get("pcid")
	ci.PartnerID = r.URL.Query().Get("pid")

	// TODO:  validate partner identity and update partner usage counter

	ci.PartnerEmailHash = r.URL.Query().Get("hem")
	ci.RedirectURL = r.URL.Query().Get("r")
	ci.ClientIP = r.Header.Get("X-Forwarded-For")
	if ci.ClientIP == "" {
		ci.ClientIP = r.Host
	}
	ci.UserAgent = r.Header.Get("User-Agent")
	cookie, err := r.Cookie(MY_COOKIE_ID)
	if err != nil {
		ci.MyCookieID = uuid.NewString()
	} else {
		ci.MyCookieID = cookie.Value
	}
	cookie = &http.Cookie{}
	cookie.Domain = x.core.Config.CookieDomain
	cookie.HttpOnly = true
	cookie.MaxAge = ONE_YEAR_SECONDS
	cookie.Name = MY_COOKIE_ID
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteNoneMode
	cookie.Secure = true
	cookie.Value = ci.MyCookieID
	http.SetCookie(w, cookie)

	log.Debug().Str("component", "monster").Str("user-agent", ci.UserAgent).Str("cookie-id", ci.MyCookieID).Str("client", ci.ClientIP).Msg("inputs")

	if ci.PartnerEmailHash == "" {
		ci.PartnerEmailHash = x.FindCookie(ci.MyCookieID).PartnerEmailHash
	}

	// sync our db
	x.SyncCookie(ci)

	// redirect is optional
	if ci.RedirectURL == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	x.Redirect(ci, w, r)
	log.Debug().Int64("microseconds", time.Now().UnixMicro()-startTime).Msg("elapsed time")
}

func (x *MonsterServer) Redirect(cm CookieInfo, w http.ResponseWriter, r *http.Request) {
	redirectURL, err := url.QueryUnescape(cm.RedirectURL)
	if err != nil {
		log.Warn().Err(err).Str("component", "moster").Str("redirect", cm.RedirectURL).Msg("query unescape")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Debug().Str("component", "monster").Str("redirect", redirectURL).Msg("redirect unescaped")

	redirectURL = strings.Replace(redirectURL, "${DEVICE_ID}", cm.MyCookieID, -1)
	redirectURL = strings.Replace(redirectURL, "${EHASH_SHA256_LOWERCASE}", cm.PartnerEmailHash, -1)

	log.Debug().Str("component", "monster").Str("redirect", redirectURL).Msg("redirect template")

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (x *MonsterServer) FindCookie(myUID string) CookieInfo {
	cm, ok := x.core.Cache.Get(myUID)
	if !ok || cm == nil {
		return CookieInfo{MyCookieID: myUID}
	}
	return cm.(CookieInfo)
}

func (x *MonsterServer) SyncCookie(newCI CookieInfo) {
	// oldCI := x.FindCookie(newCI.MyCookieID)
	// TODO:  sync cookie old/new
	x.core.Cache.Set(newCI.MyCookieID, newCI, time.Duration(ONE_YEAR_SECONDS))
	// TODO:  add to graph and consolodate graph in background
}
