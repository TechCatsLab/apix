/*
 * Revision History:
 *     Initial: 2018/06/02        Wang RiYu
 */

package main

import (
	"log"
	"time"

	"github.com/TechCatsLab/apix/geoip2"
	"github.com/TechCatsLab/apix/http/server"
	"github.com/robfig/cron"
)

// Query ...
type Query struct {
	IP string `json:"ip"`
}

// Result ...
type Result = geoip2.Result

var client = &geoip2.Client{
	DBLocationDir: "./maxminddb",
	Timeout:       time.Second * 15,
	MaxConnect:    0x64,
}

func main() {
	client.Init()
	cron := cron.New()

	cron.Start()
	defer func() {
		cron.Stop()
		client.Close()
	}()

	cron.AddFunc("0 30 2 * * 3", func() {
		client.UpdateDB()
	})

	config := &server.Configuration{Address: ":3355"}
	ep := server.NewEntrypoint(config, nil)

	router := server.NewRouter()
	router.Get("/meta", getMeta)
	router.Post("/geo", lookup)

	ep.Start(router.Handler())

	ep.Run()
}

func lookup(ctx *server.Context) error {
	var (
		resp   = ctx.Response()
		query  Query
		result *Result
	)

	if err := ctx.JSONBody(&query); err != nil {
		log.Println(err)
		return err
	}

	result, err := client.Lookup(query.IP)
	if err != nil {
		log.Println(err)
		return err
	}

	resp.WriteHeader(200)
	ctx.ServeJSON(&result)

	return nil
}

func getMeta(ctx *server.Context) error {
	res := ctx.Response()
	res.WriteHeader(200)
	res.Write([]byte("hello world! \n"))

	return nil
}
