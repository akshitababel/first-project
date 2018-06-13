package sports

import (
	"data"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"util"

	"github.com/julienschmidt/httprouter"
	"github.com/thegeniusgroup/isgdatalib"
)


// GeniusOddsFixtureList : Gets list of fixtures matching the parameters only for upcoming.
// GET  /{:sport}/{:league}
func GeniusOddsFixtureList(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	//customer := util.CustomerFromScope(p)
	typeVal := util.CleanText(p.ByName("type"), true, true)
	sportname := util.CleanText(p.ByName("sport"), true, true)
	leaguename := util.CleanText(p.ByName("league"), true, true)
	matchID := util.CleanText(p.ByName("matchid"), true, true)

	if typeVal != "best" && typeVal != "upcoming" && typeVal != "plunge" {
		util.WebResponse(w, r, http.StatusBadRequest, "Invalid type value :"+typeVal)
		return
	}

	var objsports []isg.Sport
	var objsport isg.Sport
	var objleague isg.League
	var err error

	if sportname == "all" || sportname == "" {
		for _, objsport := range data.SportObjects {
			objsports = append(objsports, objsport)
		}
	} else {
		objsport, err = data.GetSport(sportname)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "sport not found")
			return
		}
		objsports = append(objsports, objsport)
	}

	//get league detail
	if leaguename != "" {
		objleague, err = data.GetLeagueID(objsport, leaguename)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "league not found")
			return
		}
	}

	if matchID != "" {
		_, err = strconv.Atoi(matchID)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "invalid id")
			return
		}
	}

	var bestMatches []isg.OddsInfo
	var plungeMatches []isg.GeniusOddsPlunge
	if typeVal == "best" {
		bestMatches, err = data.GetGeniusOddBestMatch(objsport.SportInternalID, objleague.LeagueInternalID, typeVal)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "record not found")
			return
		} else if len(bestMatches) == 0 {
			util.WebResponse(w, r, http.StatusNotFound, "record not found")
			return
		}

	} else if typeVal == "plunge" && matchID == "" {
		plungeMatches, err = data.GetGeniusOddPlungeMatch(objsport.SportInternalID, objleague.LeagueInternalID, typeVal, matchID)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "record not found")
			return
		} else if len(plungeMatches) == 0 {
			util.WebResponse(w, r, http.StatusNotFound, "record not found")
			return
		}
	}

	// loop for sports
	var objMatch []isg.GeniusSportsMatch
	for _, objsport := range objsports {

		if objsport.SportInternalID != 1 && objsport.SportInternalID != 7 {
			continue
		}

		objLeagues := data.SportsLeagues[strconv.Itoa(objsport.SportInternalID)]

		for _, objLeague := range objLeagues {

			if leaguename != "" {
				if !util.VerifyStringInInterface(leaguename, objLeague) {
					continue
				}
			}

			if typeVal == "plunge" {

				var objmatchID []string
				var livePlungeOdds map[int64][]isg.GeniusOddsPlunge

				if matchID == "" {
					livePlungeOdds, objmatchID = isg.GetPlungeMatchID(plungeMatches, objsport.SportInternalID, objLeague.LeagueInternalID, typeVal)
				} else {

					matchid, _ := strconv.Atoi(matchID)

					plungeOddsFluc, err := data.GetMatchesProviderMarketFlucs(objsport, objLeague.LeagueInternalID, matchid, typeVal)
					if err != nil {
						fmt.Println(err.Error())
					}

					liveOdds, err := data.GetMatchesGeniusOddsPlunges(plungeOddsFluc, objsport, objLeague.LeagueInternalID, matchid)
					if err != nil {
						fmt.Println(err.Error())
						continue
					}

					if len(liveOdds) > 0 {
						livePlungeOdds = isg.MakingLiveOddsChangeSort(liveOdds, objsport.SportInternalID)
					}

					objmatchID = append(objmatchID, matchID)
				}

				objMatch, err = data.GetPlungeMatchesForGeniusOdds(objMatch, livePlungeOdds, objsport, objLeague, typeVal, objmatchID)
				if err != nil {
					continue
				}
			} else if typeVal == "best" {

				bestMatchID := isg.GetBestMatchID(bestMatches, objsport.SportInternalID, objLeague.LeagueInternalID, typeVal)

				var liveOdd []isg.IntMarketInfo

				marketOddsFluc, err := data.GetMatchesProviderMarketFlucs(objsport, objLeague.LeagueInternalID, bestMatchID, typeVal)
				if err != nil {
					fmt.Println(err.Error())
				}

				liveOdds, err := data.GetMatchesProviderMarketOdds(marketOddsFluc, objsport, objLeague.LeagueInternalID, bestMatchID, typeVal)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				if len(liveOdds) > 0 {
					liveOdd = isg.MakingGeniusLiveOddsSort(liveOdds, objsport.SportInternalID, typeVal)
				}

				_sqlstr := data.GenerateSQLQueryForGeniusOdds(objsport, objLeague, bestMatchID) // third parameter id optional parameter as matchID
				objMatch, err = data.GetMatchesForGeniusOdds(_sqlstr, objMatch, liveOdd, objsport, objLeague, typeVal)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
			} else {

				var liveOdd []isg.IntMarketInfo

				marketOddsFluc, err := data.GetMatchesProviderMarketFlucs(objsport, objLeague.LeagueInternalID, 0, typeVal)
				if err != nil {
					fmt.Println(err.Error())
				}

				liveOdds, err := data.GetMatchesProviderMarketOdds(marketOddsFluc, objsport, objLeague.LeagueInternalID, 0, typeVal)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				if len(liveOdds) > 0 {
					liveOdd = isg.MakingGeniusLiveOddsSort(liveOdds, objsport.SportInternalID, typeVal)
				}

				_sqlstr := data.GenerateSQLQueryForGeniusOdds(objsport, objLeague, 0) // third parameter id optional parameter as matchID
				objMatch, err = data.GetMatchesForGeniusOdds(_sqlstr, objMatch, liveOdd, objsport, objLeague, typeVal)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
			}

		}
	}

	if len(objMatch) == 0 {
		util.WebResponse(w, r, http.StatusNotFound, "record not found")
		return
	}

	// Sort the matches according to listing parameter
	objMatch[0].TypeVal = typeVal
	sort.Sort(isg.GeniusSortMatchesISG(objMatch))

	// if typeVal == "best" && len(objMatch) > 1 {
	// 	bestMatch := objMatch[0]
	// 	objMatch = []isg.GeniusSportsMatch{}
	// 	objMatch = append(objMatch, bestMatch)
	// }

	// Binding the matches into json
	t := isg.BindingGeniusOddsMatches(objMatch, typeVal)
	final := util.JSONMessageWrappedObj(http.StatusOK, t)
	util.WebResponseJSONObjectNoCache(w, r, http.StatusOK, final)
	return
}

// GeniusOddsMarketFixtureList :
func GeniusOddsMarketFixtureList(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err error
	sportname := util.CleanText(p.ByName("sport"), true, true)
	leaguename := util.CleanText(p.ByName("league"), true, true)
	season := util.CleanText(p.ByName("season"), true, true)
	round := util.CleanText(p.ByName("round"), true, true)
	team1 := util.CleanText(p.ByName("team1"), true, true)
	team2 := util.CleanText(p.ByName("team2"), true, true)

	typeVal := "market"

	if sportname == "" {
		util.WebResponse(w, r, http.StatusNotFound, "sport not found")
		return
	}

	objsport, err := data.GetSport(sportname)
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "sport not found "+sportname)
		return
	}

	if leaguename == "" {
		util.WebResponse(w, r, http.StatusNotFound, "league not found")
		return
	}

	objleague, err := data.GetLeagueID(objsport, leaguename)
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "league not found")
		return
	}

	if season == "" {
		util.WebResponse(w, r, http.StatusNotFound, "season not found")
		return
	}

	// change season table for cricket
	if objsport.SportInternalID == 5 && objleague.LeagueInternalID == 2 {
		objsport.TableNameSeasons = "isg_cricket_single_season"
	}

	// change season table for soccer worldcup
	if objsport.SportInternalID == 4 && objleague.LeagueInternalID == 17 {
		objsport.TableNameSeasons = "isg_soccer_worldcup_season"
	}

	// change matches table for basketball NBL
	if objsport.SportInternalID == 3 && objleague.LeagueInternalID == 2 {
		objsport.TableNameMatches = "isg_basketball_round_matches"
	}

	seasonID, err := data.GetSeasonID(objsport, season)
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "season not found")
		return
	}

	if round == "" {
		util.WebResponse(w, r, http.StatusNotFound, "date/round/week not found")
		return
	}
	var roundWeekDate string

	roundArr := strings.Split(round, "-")
	if len(roundArr) == 2 {
		checkCount, err := time.Parse("Jan", roundArr[0])
		if err == nil {
			month := checkCount.Format("01")
			roundWeekDate = "-" + month + "-" + roundArr[1]
		} else {
			objRound, _, err := data.GetRoundWeekDetails(objsport, objleague, round)
			if err != nil {
				util.WebResponse(w, r, http.StatusNotFound, "date/round/week not found")
				return
			}
			if len(objRound) > 0 {
				roundWeekDate = strconv.Itoa(objRound[0].RoundID)
			}
		}
	} else {
		objRound, _, err := data.GetRoundWeekDetails(objsport, objleague, round)
		if err != nil {
			util.WebResponse(w, r, http.StatusNotFound, "date/round/week not found")
			return
		}
		if len(objRound) > 0 {
			roundWeekDate = strconv.Itoa(objRound[0].RoundID)
		}
	}

	if team1 == "" || team2 == "" {
		util.WebResponse(w, r, http.StatusNotFound, "teams not found")
		return
	}

	sortOder := "ASC"
	if strings.HasSuffix(team2, "-2") {
		sortOder = "DESC"
		team2 = strings.Replace(team2, "-2", "", -1)
	}

	var homeTeamName, awayTeamName string
	if objsport.SportInternalID == 2 || (objsport.SportInternalID == 3 && objleague.LeagueInternalID == 1) || objsport.SportInternalID == 8 ||
		objsport.SportInternalID == 9 {
		homeTeamName = team2
		awayTeamName = team1
	} else {
		homeTeamName = team1
		awayTeamName = team2
	}

	homeTeamID, err := data.GetTeam(objsport.SportInternalID, homeTeamName)
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "home team not found")
		return
	}

	awayTeamID, err := data.GetTeam(objsport.SportInternalID, awayTeamName)
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "away team not found")
		return
	}

	matchID, err := data.GetGeniusMarketMatchID(objsport, objleague.LeagueInternalID, seasonID, roundWeekDate, homeTeamID, awayTeamID, sortOder)

	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "unable to get match record")
		return
	}

	if matchID == 0 {
		util.WebResponse(w, r, http.StatusNotFound, "match not found")
		return
	}

	var objMatch []isg.GeniusSportsMatch

	// get plunge match
	plungeMatches, _ := data.GetGeniusOddPlungeMatch(objsport.SportInternalID, objleague.LeagueInternalID, "plunge", strconv.Itoa(matchID))

	var liveOdd []isg.IntMarketInfo

	marketOddsFluc, err := data.GetMarketMatchesProviderFlucs(objsport, objleague.LeagueInternalID, matchID, typeVal)
	if err != nil {
		fmt.Println(err.Error())
	}

	liveOdds, err := data.GetMarketMatchesProviderOdds(marketOddsFluc, objsport, objleague.LeagueInternalID, matchID)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(liveOdds) > 0 {
		liveOdd = isg.MakingGeniusLiveMarketOddsSort(liveOdds, objsport.SportInternalID, "market")
	}

	_sqlstr := data.GenerateSQLQueryForGeniusOdds(objsport, objleague, matchID)
	objMatch, err = data.GetMatchesForGeniusOdds(_sqlstr, objMatch, liveOdd, objsport, objleague, "market")
	if err != nil {
		util.WebResponse(w, r, http.StatusNotFound, "unable to get the match record.")
		return
	}
	if len(objMatch) == 0 {
		util.WebResponse(w, r, http.StatusNotFound, "record not found")
		return
	}

	t := isg.BindingGeniusOddsMarketMatches(objMatch, typeVal, plungeMatches)
	final := util.JSONMessageWrappedObj(http.StatusOK, t)
	util.WebResponseJSONObjectNoCache(w, r, http.StatusOK, final)
	return
}
