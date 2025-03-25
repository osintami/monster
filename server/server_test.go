// © 2022 Sloan Childers
package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/osintami/monster/utils"
	"github.com/osintami/plumbr/sink"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCookieSyncInCache(t *testing.T) {
	router, _, config := InitServer(t)
	ci := InitCookieInfo(t)
	//cache.On("Get", mock.Anything).Return(ci, true)

	w := httptest.NewRecorder()
	path := fmt.Sprintf("/csr?pcid=%s&pid=%s&hem=%s&r=%s", ci.PartnerCookieID, ci.PartnerID, ci.PartnerEmailHash, ci.RedirectURL)
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Add("User-Agent", "test-user-agent")
	req.Header.Add("X-Forwarded-For", "220.120.12.13")
	req.AddCookie(&http.Cookie{Name: MY_COOKIE_ID, Value: ci.MyCookieID})

	router.ServeHTTP(w, req)

	assert.Equal(t, 302, w.Code)
	resp := w.Result()
	location, _ := resp.Header["Location"]
	assert.Equal(t, ci.RedirectURL, location[0])

	cookies, _ := resp.Header["Set-Cookie"]
	expectedCookie := fmt.Sprintf("muid=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly; Secure; SameSite=None", ci.MyCookieID, config.CookieDomain, ONE_YEAR_SECONDS)
	assert.Equal(t, expectedCookie, cookies[0])
}

func TestCookieSyncRedirectTemplate(t *testing.T) {
	router, cache, config := InitServer(t)
	ci := InitCookieInfo(t)
	ci.PartnerEmailHash = ""
	ci.RedirectURL = url.QueryEscape("https://google.com/?uid=${DEVICE_ID}&hem=${EHASH_SHA256_LOWERCASE}")
	cache.On("Get", mock.Anything).Return(ci, false)

	w := httptest.NewRecorder()
	path := fmt.Sprintf("/csr?pcid=%s&pid=%s&r=%s", ci.PartnerCookieID, ci.PartnerID, ci.RedirectURL)
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Add("User-Agent", "test-user-agent")
	req.Header.Add("X-Forwarded-For", "220.120.12.13")
	req.AddCookie(&http.Cookie{Name: MY_COOKIE_ID, Value: ci.MyCookieID})

	router.ServeHTTP(w, req)

	assert.Equal(t, 302, w.Code)
	resp := w.Result()
	location, _ := resp.Header["Location"]
	expectedURL := fmt.Sprintf("https://google.com/?uid=%s&hem=", ci.MyCookieID)
	assert.Equal(t, expectedURL, location[0])

	cookies, _ := resp.Header["Set-Cookie"]
	expectedCookie := fmt.Sprintf("muid=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly; Secure; SameSite=None", ci.MyCookieID, config.CookieDomain, ONE_YEAR_SECONDS)
	assert.Equal(t, expectedCookie, cookies[0])
}

func TestCookieSyncNoRedirect(t *testing.T) {
	router, cache, config := InitServer(t)
	ci := InitCookieInfo(t)
	ci.RedirectURL = ""
	ci.PartnerEmailHash = ""
	cache.On("Get", mock.Anything).Return(ci, true)

	w := httptest.NewRecorder()

	path := fmt.Sprintf("/csr?pcid=%s&pid=%s", ci.PartnerCookieID, ci.PartnerID)
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Add("User-Agent", "test-user-agent")
	req.Header.Add("X-Forwarded-For", "220.120.12.13")
	req.AddCookie(&http.Cookie{Name: MY_COOKIE_ID, Value: ci.MyCookieID})

	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	resp := w.Result()

	location, _ := resp.Header["Location"]
	assert.Equal(t, 0, len(location))

	cookies, _ := resp.Header["Set-Cookie"]
	expectedCookie := fmt.Sprintf("muid=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly; Secure; SameSite=None", ci.MyCookieID, config.CookieDomain, ONE_YEAR_SECONDS)
	assert.Equal(t, expectedCookie, cookies[0])
}

func InitServer(t *testing.T) (*chi.Mux, *MockCache, utils.ServerConfig) {
	cache := NewMockCache(t)
	cfg := utils.ServerConfig{CookieDomain: "a.osintami.com", PathPrefix: "/", LogLevel: "trace"}
	sink.InitLogger(cfg.LogLevel)
	core := utils.ServerCore{
		Config: cfg,
		Cache:  cache,
	}
	in := NewServer(core)

	router := chi.NewMux()
	router.Route(core.Config.PathPrefix, func(r chi.Router) {
		r.Get("/csr", in.CookieSync)
	})

	return router, cache, cfg
}

func InitCookieInfo(t *testing.T) CookieInfo {
	return CookieInfo{
		MyCookieID:       "test-my-cookie-id",
		PartnerCookieID:  "test-partner-cookie-id",
		PartnerID:        "test-partner-id",
		PartnerEmailHash: "test-email-hash",
		RedirectURL:      "/some-random-path"}
}

// NOTE:  built by hand to work with Mock, modify at your own peril
type IMockCache interface {
	mock.TestingT
	Cleanup(func())
}

type MockCache struct {
	mock.Mock
}

func NewMockCache(t IMockCache) *MockCache {
	mock := &MockCache{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}

func (x *MockCache) Get(key string) (interface{}, bool) {
	ret := x.Called(key)

	var r0 CookieInfo
	if rf, ok := ret.Get(0).(func(string) CookieInfo); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(CookieInfo)
	}

	return r0, true
}

func (x *MockCache) Set(key string, value interface{}, d time.Duration) {
	return
}
