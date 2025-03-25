// © 2022 Sloan Childers
package server

import (
	"context"
	"testing"

	"github.com/osintami/monster/utils"
	"github.com/stretchr/testify/assert"
)

func TestSaveCookie(t *testing.T) {
	dg, ctx := InitDgraph(t)
	txn := dg.NewTxn()

	cookie := &utils.Cookie{
		Uid:      "_:cookie",
		CookieID: "xyz123",
		Browsers: []utils.Browser{{
			Uid:       "_:browser",
			Addr:      "220.120.12.13",
			UserAgent: "test-user-agent",
			Count:     45,
		}},
		Partners: []utils.Partner{{
			Uid:       "_:partner",
			PartnerID: "pdq123",
			CookieID:  "xyz456",
		}},
	}

	cookie, err := dg.CreateCookie(ctx, txn, cookie, false)
	assert.NoError(t, err)
	assert.Equal(t, "xyz123", cookie.CookieID)

	//	dg.DeleteCookie(ctx, nil, cookie, true)
	cookie, err = dg.FindCookieByUid(ctx, txn, cookie.Uid)
	assert.Nil(t, err)
	assert.NotNil(t, cookie)

	cookie, err = dg.FindCookie(ctx, txn, cookie.CookieID)
	assert.Nil(t, err)
	assert.NotNil(t, cookie)

	txn.Discard(ctx)
}

func TestDeleteCookie(t *testing.T) {
	dg, ctx := InitDgraph(t)
	cookie, err := dg.FindCookie(ctx, nil, "xyz123")
	assert.Equal(t, utils.ErrCookieNotFound, err)
	err = dg.DeleteCookie(ctx, nil, cookie, true)
	assert.Nil(t, err)
}

func TestSaveBrowser(t *testing.T) {
	dg, ctx := InitDgraph(t)
	txn := dg.NewTxn()

	browser := &utils.Browser{
		Uid:       "_:browser",
		Addr:      "220.120.12.13",
		UserAgent: "test-user-agent",
		Count:     45,
	}

	browser, err := dg.CreateBrowser(ctx, txn, browser, false)
	assert.NoError(t, err)
	assert.Equal(t, "220.120.12.13", browser.Addr)

	//	dg.DeleteCookie(ctx, nil, cookie, true)
	browser, err = dg.FindBrowserByUid(ctx, txn, browser.Uid)
	assert.Nil(t, err)
	assert.NotNil(t, browser)

	browser, err = dg.FindBrowser(ctx, txn, "test-user-agent", "220.120.12.13")
	assert.Nil(t, err)
	assert.NotNil(t, browser)

	txn.Discard(ctx)
}

func TestCollapseBrowser(t *testing.T) {
	// TODO:  make a Cookie, check for existing browser
	//   and use the existing one with the new Cookie
	dg, ctx := InitDgraph(t)
	txn := dg.NewTxn()

	// browser exists, but with some other cookie
	cookie1 := &utils.Cookie{
		Uid:      "_:cookie",
		CookieID: "xyz123",
		Browsers: []utils.Browser{{
			Uid:       "_:browser",
			Addr:      "220.120.12.13",
			UserAgent: "test-user-agent",
			Count:     45,
		}},
		Partners: []utils.Partner{{
			Uid:       "_:partner",
			PartnerID: "pdq123",
			CookieID:  "xyz456",
		}},
	}

	cookie1, err := dg.CreateCookie(ctx, txn, cookie1, false)
	assert.NoError(t, err)
	assert.Equal(t, "xyz123", cookie1.CookieID)

	// pixel fire...  cookie does not exist, existing partner
	// TODO:  code up
	partners := []utils.Partner{{
		Uid:       "_:partner",
		PartnerID: "pdq123",
		CookieID:  "xyz456",
	}}

	_, err = dg.FindCookie(ctx, txn, "pdq123")
	if err == utils.ErrCookieNotFound {
		browser, err := dg.FindBrowser(ctx, txn, "test-user-agent", "220.120.12.13")
		if err == utils.ErrBrowserNotFound {
			// create new browser
		}
		cookie2 := &utils.Cookie{
			Uid:      "_:cookie",
			CookieID: "xyz456",
			Browsers: []utils.Browser{*browser},
			Partners: partners,
		}

		cookie2, err = dg.CreateCookie(ctx, txn, cookie2, false)
		assert.NoError(t, err)
		assert.Equal(t, "xyz456", cookie2.CookieID)
	}

}

// func TestNewCookie(t *testing.T) {

// }

// func TestOldCookie(t *testing.T) {

// }

func InitDgraph(t *testing.T) (*utils.Dgraph, context.Context) {
	cfg := utils.ServerConfig{DgraphSvr: "localhost:9080"}
	dg := utils.NewDgraph(cfg)
	ctx := context.Background()
	err := dg.DropSchema(ctx)
	assert.NoError(t, err)
	err = dg.CreateSchema(ctx)
	assert.NoError(t, err)
	return dg, ctx
}
