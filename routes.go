package main

import (
	"api/analysis"
	"api/funfacts"
	"api/geniusodds"
	"api/portal"
	"api/predictor"
	"api/sports"
	"api/stats"
	"api/tips"
	"api/trending"
	"api/user"
	"api/weather"
	"api/widgets"
	"fmt"
	"net/http"
	"util"

	"github.com/julienschmidt/httprouter"
)

	router.GET("/geniusodds/matches/:type", sports.GeniusOddsFixtureList)
	router.GET("/geniusodds/matches/:type/:sport", sports.GeniusOddsFixtureList)
	router.GET("/geniusodds/matches/:type/:sport/:league", sports.GeniusOddsFixtureList)
	router.GET("/geniusodds/matches/:type/:sport/:league/:matchid", sports.GeniusOddsFixtureList)
	router.GET("/geniusodds/markets/:sport/:league/:season/:round/:team1/:team2", sports.GeniusOddsMarketFixtureList)

	