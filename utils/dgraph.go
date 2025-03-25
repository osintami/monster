// © 2022 Sloan Childers
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/dgo/v2"
	api "github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type CookieResponse struct {
	All []Cookie `json:"all"`
}

type BrowserResponse struct {
	All []Browser `json:"all"`
}

type Browser struct {
	Uid       string `json:"uid,omitempty"`
	Addr      string `json:"addr"`
	UserAgent string `json:"useragent"`
	Count     int    `json:"count"`
}
type Partner struct {
	Uid       string `json:"uid,omitempty"`
	PartnerID string `json:"pid"`
	CookieID  string `json:"pcookie"`
}
type Cookie struct {
	Uid      string     `json:"uid,omitempty"`
	CookieID string     `json:"cookie"`
	IssuedAt *time.Time `json:"issued"`
	Browsers []Browser  `json:"browser"`
	Partners []Partner  `json:"partner"`
}

type Dgraph struct {
	dg *dgo.Dgraph
}

var ErrCookieNotFound = errors.New("cookie not found")
var ErrBrowserNotFound = errors.New("browser not found")

var ErrDuplicateCookiesExist = errors.New("duplicate cookies exist")

func NewDgraph(cfg ServerConfig) *Dgraph {
	conn, err := grpc.Dial(cfg.DgraphSvr, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Err(err).Msg("connect")
	}
	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	return &Dgraph{dg: dg}
}

func (x *Dgraph) DropSchema(ctx context.Context) error {
	return x.dg.Alter(ctx, &api.Operation{DropOp: api.Operation_ALL})
}

func (x *Dgraph) DropData(ctx context.Context) error {
	return x.dg.Alter(context.Background(), &api.Operation{DropOp: api.Operation_DATA})
}

func (x *Dgraph) CreateSchema(ctx context.Context) error {

	op := &api.Operation{}
	op.Schema = `
			cookie: string @index(hash) .
			issued: datetime .
			Browser: [uid] .
			Partner: [uid] .
			addr: string @index(hash) .
			useragent: string @index(hash) .
			count: int .
			pid: string .
			pcookie: string @index(hash) .
	
			type Browser {
				addr: string
				useragent: string
				count: int
			}		
			type Partner {
				pid: string!
				pcookie: string!
			}
			type Cookie {
				cookie: string! 
				issued: datetime
				Browser: [Browser]
				Partner: [Partner]
			}	
		`

	if err := x.dg.Alter(ctx, op); err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("schema")
		return err
	}
	return nil
}

func (x *Dgraph) CreateBrowser(ctx context.Context, txn *dgo.Txn, browser *Browser, commitNow bool) (*Browser, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}

	mu := &api.Mutation{}
	pb, err := json.Marshal(browser)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "create").Msg("marshal")
		return browser, err
	}

	mu.SetJson = pb
	if commitNow {
		mu.CommitNow = true
	}

	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "create").Msg("mutate")
		return browser, err
	}

	uid := resp.GetUids()["browser"]
	return x.FindBrowserByUid(ctx, txn, uid)
}

func (x *Dgraph) CreateCookie(ctx context.Context, txn *dgo.Txn, cookie *Cookie, commitNow bool) (*Cookie, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}

	createdAt := time.Now()
	cookie.IssuedAt = &createdAt

	mu := &api.Mutation{}
	pb, err := json.Marshal(cookie)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "create").Msg("marshal")
		return cookie, err
	}

	mu.SetJson = pb
	if commitNow {
		mu.CommitNow = true
	}

	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "create").Msg("mutate")
		return cookie, err
	}

	uid := resp.GetUids()["cookie"]
	return x.FindCookieByUid(ctx, txn, uid)
}

func (x *Dgraph) DeleteCookie(ctx context.Context, txn *dgo.Txn, cookie *Cookie, commitNow bool) error {

	if txn == nil {
		txn = x.dg.NewTxn()
	}

	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(cookie)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "delete").Msg("marshal")
		return err
	}

	mu.DeleteJson = pb
	if commitNow {
		mu.CommitNow = true
	}

	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Str("command", "delete").Msg("mutate")
		return err
	}

	return err
}

func (x *Dgraph) NewTxn() *dgo.Txn {
	return x.dg.NewTxn()
}

func (x *Dgraph) FindCookie(ctx context.Context, txn *dgo.Txn, cookie string) (*Cookie, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}
	vars := map[string]string{"$cookie": cookie}
	query := `query all($cookie: string) {
		all(func: eq(cookie, $cookie)) {
			uid
			cookie
			issued
			browser {
				uid
				addr
				useragent
				count
			}
			partner {
				uid
				pid
				pcookie
			}
		}
	}
	`
	resp, err := txn.QueryWithVars(ctx, query, vars)
	//resp, err := txn.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("find cookie")
		return nil, err
	}
	return x.returnCookie(resp.Json)
}

func (x *Dgraph) FindBrowser(ctx context.Context, txn *dgo.Txn, ua string, ip string) (*Browser, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}
	vars := make(map[string]string)
	vars["$ua"] = ua
	vars["$ip"] = ip
	query := `query all($ua: string, $ip: string) {
		all(func: eq(useragent, $ua)) @filter(eq(addr, $ip)) {
			uid
			addr
			useragent
			count
		}
	}
	`
	resp, err := txn.QueryWithVars(ctx, query, vars)
	//resp, err := txn.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("find cookie")
		return nil, err
	}
	return x.returnBrowser(resp.Json)
}

func (x *Dgraph) FindBrowserByUid(ctx context.Context, txn *dgo.Txn, uid string) (*Browser, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}

	vars := map[string]string{"$uid": uid}
	query := `
		query browsers($uid: string) {
			all(func: uid($uid)) {
				uid
				addr
				useragent
				count
			}
		}
		`

	resp, err := txn.QueryWithVars(ctx, query, vars)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("find browser with uid")
		return nil, err
	}
	return x.returnBrowser(resp.Json)
}

func (x *Dgraph) FindCookieByUid(ctx context.Context, txn *dgo.Txn, uid string) (*Cookie, error) {

	if txn == nil {
		txn = x.dg.NewTxn()
	}

	vars := map[string]string{"$uid": uid}
	query := `
		query cookies($uid: string) {
			all(func: uid($uid)) {
				uid
				cookie
				issued
				browser {
					uid
					addr
					useragent
					count
				}
				partner {
					uid
					pid
					pcookie
				}
			}
		}
		`

	resp, err := txn.QueryWithVars(ctx, query, vars)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("find cookie with uid")
		return nil, err
	}
	return x.returnCookie(resp.Json)
}

func (x *Dgraph) returnCookie(resp []byte) (*Cookie, error) {

	log.Debug().Str("component", "dgraph").Str("json", string(resp)).Msg("result cookie")

	var data CookieResponse
	err := json.Unmarshal(resp, &data)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("unmarshal")
		return nil, err
	}

	if len(data.All) == 0 {
		return nil, ErrCookieNotFound
	}

	if len(data.All) > 1 {
		return &data.All[0], ErrDuplicateCookiesExist
	}

	return &data.All[0], nil
}

func (x *Dgraph) returnBrowser(resp []byte) (*Browser, error) {

	log.Debug().Str("component", "dgraph").Str("json", string(resp)).Msg("result browser")

	var data BrowserResponse
	err := json.Unmarshal(resp, &data)
	if err != nil {
		log.Error().Err(err).Str("component", "dgraph").Msg("unmarshal")
		return nil, err
	}

	if len(data.All) == 0 {
		return nil, ErrCookieNotFound
	}

	return &data.All[0], nil
}
