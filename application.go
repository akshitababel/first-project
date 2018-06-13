package main

import (
	"data"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/thegeniusgroup/isgdatalib"
)

func main() {

	var port = os.Getenv("RDS_PORT")
	fmt.Println(port)

	dbHost := "localhost"
	dbPort := "3306"
	dbUser := "root"
	dbPass := ""
	dbHostAU := "localhost"
	dbPassAU := ""
	cacheDataHost := "localhost"
	cacheDataPort := "6379"
	cacheLogHost := "localhost"
	cacheLogPort := "6379"
	authCacheHost := "localhost"
	authCachePort := "6379"

	var env = strings.ToLower(os.Getenv("Environment"))
	if env == "production" {
		dbHost = os.Getenv("RDS_HOSTNAME")
		dbUser = os.Getenv("RDS_USERNAME")
		dbPort = os.Getenv("RDS_PORT")
		dbPass = os.Getenv("RDS_PASSWORD")
		cacheDataHost = os.Getenv("CACHE_RW_HOST")
		cacheDataPort = os.Getenv("CACHE_RW_PORT")
		cacheLogHost = os.Getenv("CACHE_LOG_HOST")
		cacheLogPort = os.Getenv("CACHE_LOG_PORT")
		authCacheHost = os.Getenv("AUTH_CACHE_HOST")
		authCachePort = os.Getenv("AUTH_CACHE_PORT")
		dbHostAU = os.Getenv("RDS_AU_DB_HOSTNAME")
		dbPassAU = os.Getenv("RDS_AU_DB_PASSWORD")
	} else if env == "development" {
		dbHost = os.Getenv("RDS_HOSTNAME")
		dbUser = os.Getenv("RDS_USERNAME")
		dbPort = os.Getenv("RDS_PORT")
		dbPass = os.Getenv("RDS_PASSWORD")
		cacheDataHost = os.Getenv("CACHE_RW_HOST")
		cacheDataPort = os.Getenv("CACHE_RW_PORT")
		cacheLogHost = os.Getenv("CACHE_LOG_HOST")
		cacheLogPort = os.Getenv("CACHE_LOG_PORT")
		authCacheHost = os.Getenv("AUTH_CACHE_HOST")
		authCachePort = os.Getenv("AUTH_CACHE_PORT")
		dbHostAU = os.Getenv("RDS_AU_DB_HOSTNAME")
		dbPassAU = os.Getenv("RDS_AU_DB_PASSWORD")
	}

	// Connect the database
	_, _, _, _, _, err := data.InitDB(dbHost, dbPort, dbUser, dbPass, dbHostAU, dbPassAU)
	if err != nil {
		log.Panic(err)
	}

	// Connect the cache server
	err = data.InitCache(cacheDataHost, cacheDataPort, cacheLogHost, cacheLogPort, authCacheHost, authCachePort)
	if err != nil {
		log.Panic(err)
	}

	// Preload Sports and Leagues for quick lookup
	preloadCachedLookupData()

	router := httprouter.New()
	router.RedirectTrailingSlash = true
	addRouteHandlers(router)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "Authorization"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		//Debug:            true,
	})

	log.Fatal(http.ListenAndServe(":3000", c.Handler(router)))
}

// preloadCachedLookupData preloads the lookup data structures (from Package data) within data/cachedata.go.
func preloadCachedLookupData() {
	preloadSports()
	preloadLeagues()
	preloadSeasons()
	preloadMarkets()
	preloadFilters()
	preloadProviders()
	preloadCustomers()
}

// Preload Sports
func preloadSports() {
	rows, err := data.SportsDb.Query("SELECT sport_api_altname, sport_api_code, sport_name, sport_id, sport_season_tablename, sport_match_tablename, sport_player_tablename,sport_url, sport_logo FROM isg_sports ORDER BY sport_id")
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	// SportAPIIDs[0] = "", SportAPIIDs[1] = "ar", ....
	data.SportAPIIDs = append(data.SportAPIIDs, "")

	for rows.Next() {
		var sport isg.Sport
		var sportAltName string

		err := rows.Scan(&sportAltName, &sport.SportID, &sport.SportName, &sport.SportInternalID, &sport.TableNameSeasons, &sport.TableNameMatches, &sport.TableNamePlayers, &sport.SportURL, &sport.SportLogo)
		if err != nil {
			log.Panic(err)
		}
		sport.SportID = strings.ToLower(sport.SportID)
		data.SportAPIIDs = append(data.SportAPIIDs, strings.ToLower(sport.SportID))
		data.SportIDForName[sportAltName] = strings.ToLower(sport.SportID) // aussie-rules -> "ar""
		data.ValidSportIDs[sport.SportID] = strings.ToLower(sport.SportID) // for quick validation of IDs
		data.SportObjects[sport.SportID] = sport
	}

	// fmt.Println(data.SportAPIIDs)
	// fmt.Println(data.SportIDForName)
	// fmt.Println(data.ValidSportIDs)
	// fmt.Println(data.SportObjects)
	fmt.Println("Sports preloaded.")
}

// Preload Leagues
func preloadLeagues() {
	rows, err := data.SportsDb.Query("SELECT sport_id, local_id, api_league_id, entity_api_id, entity_name FROM isg_api_entities WHERE entity_type = 'league' ORDER BY sport_id, local_id")
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	var sportLeagues []isg.League
	var league isg.League
	tmpSportId := "1"
	rowSportId := "1"

	for rows.Next() {
		err := rows.Scan(&rowSportId, &league.LeagueInternalID, &league.LeagueID, &league.LeagueEntityKey, &league.LeagueName)
		if err != nil {
			log.Panic(err)
		}
		data.LeagueIDs[league.LeagueEntityKey] = league.LeagueID

		if rowSportId != tmpSportId {
			// Finish all leagues of one sport - Assign league array to map and make a new slice
			tempLeagues := make([]isg.League, len(sportLeagues))
			copy(tempLeagues, sportLeagues)
			data.SportsLeagues[tmpSportId] = tempLeagues
			sportLeagues = []isg.League{}
			tmpSportId = rowSportId // Start counting leagues for the next sport
		}
		sportLeagues = append(sportLeagues, league)
	}
	data.SportsLeagues[tmpSportId] = sportLeagues // Assign the last batch of leagues at end of Rows

	//	if err := stats.InitStatsTennis(); err != nil {
	//		log.Panic(err)
	//	}

	// fmt.Println(data.LeagueIDs)
	//fmt.Println(data.SportsLeagues)
	// fmt.Println("Rugby League: ")
	// fmt.Println(data.SportsLeagues["7"])
	fmt.Println("Leagues preloaded.")

}

func preloadSeasons() {

}

func preloadMarkets() {

	// for _, s := range data.SportObjects {
	// 	for _, l := range data.SportsLeagues[strconv.Itoa(s.SportInternalID)] {
	// 		markets, err := data.GetMarkets(s, l)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		//marketmap := make(map[string][]isg.PortalMarketResults)
	// 		//marketmap[l.LeagueID] = markets
	// 		if len(markets) > 0 {
	// 			data.SportsMarkets[s][l.LeagueID] = markets
	// 		}

	// 	}
	// }

	// fmt.Println(data.SportsMarkets[data.SportObjects["rl"]])
	// fmt.Println(data.SportsMarkets[data.SportObjects["rl"]]["01"])

	// fmt.Println("Markets preloaded.")
}

func preloadFilters() {

}

// Preload data providers
func preloadProviders() {
	rows, err := data.SportsDb.Query("SELECT provider_id, provider_name, provider_url, provider_icon, is_td_provider, for_predictor, genius_odds_sequence FROM isg_providers WHERE status = 1 ")
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	rowData := data.StringArray(rows)

	for _, p := range rowData {
		provider := isg.Provider{}
		provider.ProviderId = p[0]
		provider.Name = p[1]
		provider.URL = p[2]
		provider.Icon = p[3]
		if p[4] == "1" {
			provider.IsTDProvider = true
		} else {
			provider.IsTDProvider = false
		}
		if p[5] == "1" {
			provider.ForPredictor = true
		} else {
			provider.ForPredictor = false
		}
		provider.GeniusOddsSequence, _ = strconv.Atoi(p[6])

		data.Providers[provider.URL] = provider
	}

	fmt.Println("Providers preloaded.")

}

// Preload customers
func preloadCustomers() {
	rows, err := data.SportsDb.Query("SELECT customer_id, api_scope_id, customer_uuid FROM isports_users.tblform_customers ")
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	rowData := data.StringArray(rows)

	for _, p := range rowData {
		customer := isg.Customer{}
		customer.Id, _ = strconv.Atoi(p[0])
		customer.Name = p[1]
		customer.UUID = p[2]
		data.Customers[p[0]] = customer
	}

	rows2, err := data.SportsDb.Query("SELECT product_id, product_name, product_uuid, provider_id, product_icon, product_url FROM isports_users.isg_clients_products")
	if err != nil {
		log.Panic(err)
	}
	defer rows2.Close()

	rowData2 := data.StringArray(rows2)

	for _, p := range rowData2 {
		product := isg.ProductInfo{}
		product.ID = p[0]
		product.ProductName = p[1]
		product.UUID = p[2]
		product.ProviderID = p[3]
		product.Icon = p[4]
		product.ProductURL = p[5]
		data.ProductInfo[product.ID] = product
	}

	fmt.Println("Customers/Products preloaded.")

}
