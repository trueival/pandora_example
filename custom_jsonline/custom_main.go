// Copyright (c) 2017 Yandex LLC. All rights reserved.
// Use of this source code is governed by a MPL 2.0
// license that can be found in the LICENSE file.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/spf13/afero"
	"github.com/yandex/pandora/cli"
	"github.com/yandex/pandora/components/phttp/import"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregator/netsample"
	"github.com/yandex/pandora/core/import"
	"github.com/yandex/pandora/core/register"
)

type Ammo struct {
	Tag         string
        Param1      string
	Param2      string
	Param3      string
}

type Sample struct {
	URL              string
	ShootTimeSeconds float64
}

type GunConfig struct {
	Target string `validate:"required"` // Configuration will fail, without target defined
}

type Gun struct {
	// Configured on construction.
	client http.Client
	conf   GunConfig
	// Configured on Bind, before shooting
	aggr core.Aggregator // May be your custom Aggregator.
	core.GunDeps
}

func NewGun(conf GunConfig) *Gun {
	return &Gun{conf: conf}
}

func (g *Gun) Bind(aggr core.Aggregator, deps core.GunDeps) error {
	g.client = http.Client{} //keep-alive shooting
	g.aggr = aggr
	g.GunDeps = deps
	return nil
}

func (g *Gun) Shoot(ammo core.Ammo) {
	customAmmo := ammo.(*Ammo) // Shoot will panic on unexpected ammo type. Panic cancels shooting.
	g.shoot(customAmmo)
}

func (g *Gun) shoot(ammo *Ammo) {
	expire := time.Now().AddDate(0, 0, 1)
	code := 0
	var uri string
	var assert string
	var method string
	switch ammo.Tag {
	case "case1":
		uri = "/case1" + ammo.Param1
		assert = "^.\"error\":"
		method = "GET"
	case "case2":
		uri = "/case2"
		assert = "^.\"error\":"
		method = "GET"
	default:
		uri = ""
		assert = ""
	}

	req, err := http.NewRequest(method, strings.Join([]string{"http://", g.conf.Target, uri}, ""), nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Connection", "close")
	req.Header.Add("x-real-ip", "127.0.0.1")
        req.AddCookie(&http.Cookie{"param2", ammo.Param2, "/", ".host.ru", expire, expire.Format(time.UnixDate), 86400, false, false, strings.Join([]string{"param2=", ammo.Param2}, ""), []string{"param2=", ammo.Param2}})
        req.AddCookie(&http.Cookie{"param3", ammo.Param3, "/", ".host.ru", expire, expire.Format(time.UnixDate), 86400, false, false, strings.Join([]string{"param3=", ammo.Param3}, ""), []string{"param3=", ammo.Param3}})

	sample := netsample.Acquire(ammo.Tag)

	rs, err := g.client.Do(req)
	if err != nil {
		code = 0
	} else {
		code = rs.StatusCode
		if code == 200 {
			respBody, _ := ioutil.ReadAll(rs.Body)
			re := regexp.MustCompile(assert)
			if re.FindString(string(respBody)) == "" || assert == "" {
				code = rs.StatusCode
			} else {
				code = 314
			}
		}
		rs.Body.Close()
	}
	defer func() {
		sample.SetProtoCode(code)
		g.aggr.Report(sample)
	}()
}

func main() {
	//debug.SetGCPercent(-1)
	// Standard imports.
	fs := afero.NewOsFs()
	coreimport.Import(fs)
	// May not be imported, if you don't need http guns and etc.
	phttp.Import(fs)

	// Custom imports. Integrate your custom types into configuration system.
	coreimport.RegisterCustomJSONProvider("custom_provider", func() core.Ammo { return &Ammo{} })

	register.Gun("custom_gun", NewGun, func() GunConfig {
		return GunConfig{
			Target: "default target",
		}
	})

	cli.Run()
}
