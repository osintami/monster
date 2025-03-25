// © 2022 Sloan Childers
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/osintami/monster/utils"
	"github.com/rs/zerolog/log"
)

// TODO:  wrap this into a dgraph class and come up with several cookie scenarios to test
//   basic cookie ops and merge ops

// TODO:  ask if non-humans are an issue (cloud nodes, bots)

// TODO:  how is GEO used from ipinfo

// TODO:  ask about the hem and when it's sent, only on login?  is it a trigger to merge
//   cookies?

// TODO:  access control, only via private ID?

// TODO:  reporting, cookie management, tell folks what cookies they can collapse?

func main() {
	cfg := utils.ServerConfig{DgraphSvr: "localhost:9080", DgraphUser: "", DgraphPass: ""}
	dg := utils.NewDgraph(cfg)

	ctx := context.Background()

	// err := dg.CreateSchema(ctx)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("schema")
	// }
	// createdAt := time.Now()
	// cookie := &utils.Cookie{
	// 	Uid:    "_:cookie",
	// 	Issued: &createdAt,
	// 	Id:     "xyz123",
	// 	Browsers: []utils.Browser{{
	// 		Uid:       "_:browser",
	// 		Addr:      "220.120.12.13",
	// 		UserAgent: "test-user-agent",
	// 		Count:     45,
	// 	}},
	// 	Partners: []utils.Partner{{
	// 		Uid:    "_:partner",
	// 		Id:     "pdq123",
	// 		Cookie: "xyz456",
	// 	}},
	// }

	// cookie, err = dg.CreateCookie(ctx, cookie)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("create")
	// }
	// out, _ := json.MarshalIndent(cookie, "", "  ")
	// fmt.Printf("%s\n", out)

	cookie, err := dg.FindCookie(ctx, nil, "xyz123")
	if err != nil {
		log.Fatal().Err(err).Msg("findbyid")
	}

	out, _ := json.MarshalIndent(cookie, "", "  ")
	fmt.Printf("%s\n", out)

}
