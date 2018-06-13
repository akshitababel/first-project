/*
Package data - Handles functions related to data source access e.g. cache, databases
*/
package data

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/shopspring/decimal"
	"github.com/thegeniusgroup/isgdatalib"
)

func GetAllSports() ([]*isg.Sport, error) {
	rows, err := SportsDb.Query("SELECT sport_id,sport_name,sport_api_code FROM isg_sports")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sports := make([]*isg.Sport, 0)
	for rows.Next() {
		sport := new(isg.Sport)
		err := rows.Scan(&sport.SportInternalID, &sport.SportName, &sport.SportID)
		if err != nil {
			return nil, err
		}
		sports = append(sports, sport)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sports, nil
}

// GetAllTeams retrieves all teams from a sport + league.
// league is expected to be in the nominal 2-char ID format i.e. 01, 02... etc...
func GetAllTeams(sport isg.Sport, league isg.League, seasons string, inclusions int, customers ...int) ([]isg.Team, error) {

	teams := []isg.Team{}
	seasonIDs, seasonNames := []string{}, []string{}
	var result string

	//making a season string for particuler team
	if seasons != "" {
		seasonNames = strings.Split(seasons, "+")
	}
	objSport := SportObjects[strings.ToLower(sport.SportID)]
	if objSport.SportID == "cr" && league.LeagueInternalID == 2 {
		objSport.TableNameSeasons = "isg_cricket_single_season"
	}
	if objSport.SportID == "sc" && (league.LeagueInternalID == 19 || league.LeagueInternalID == 24 || league.LeagueInternalID == 25 || league.LeagueInternalID == 26 || league.LeagueInternalID == 27 || league.LeagueInternalID == 28 || league.LeagueInternalID == 29) {
		objSport.TableNameSeasons = "isg_sports_league_seasons"
	}
	if len(seasonNames) > 0 {
		for _, s := range seasonNames {
			id, _ := GetSeasonID(objSport, s)
			seasonIDs = append(seasonIDs, strconv.Itoa(id))
		}
		sort.Strings(seasonIDs)
		result = strings.Join(seasonIDs, ",")
	}
	// Check cached data first, use it if available
	/*	var cacheData []byte
		cacheData, err := CacheGet("isgapi:obj:teamlist:" + sport.SportID + ":" + league.LeagueID)
		if err == nil && cacheData != nil {
			returnData := isg.Extract(cacheData)
			err = json.Unmarshal(returnData, &teams)
			return teams, nil
		}*/

	var homeSQL, awaySQL, teamSQL, rankSQL, customerSQL string
	switch strings.ToLower(sport.SportID) {
	case "sc":
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND week_id >= 1 AND week_id <= 52 "
		}
		homeSQL = "SELECT home_team_id FROM isg_soccermatches where league_id = " + strconv.Itoa(league.LeagueInternalID)
		awaySQL = "SELECT away_team_id FROM isg_soccermatches where league_id = " + strconv.Itoa(league.LeagueInternalID)
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT week_id from isg_soccermatches where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}

	case "ar":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches

		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND round_id >= 1 AND round_id <= 24 "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND round_id >= 25 AND round_id <= 29 "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT round_id from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"

		}
	case "ru":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND round_id >= 1 AND round_id <= 20 "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND round_id >= 21 AND round_id <= 23 "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT round_id from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}
	case "rl":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND round_id >= 1 AND round_id <= 26 "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND round_id >= 27 AND round_id <= 30 "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT round_id from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}
	case "af":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND week_id >= 1 AND week_id <= 18 "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND week_id >= 19 AND week_id <= 22 "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT week_id from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}
	case "ih":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND week_id = 1  "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND week_id >= 2 AND week_id <= 8 "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT week_id from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}
	case "bl":
		homeSQL = "SELECT home_team_id FROM " + sport.TableNameMatches
		awaySQL = "SELECT away_team_id FROM " + sport.TableNameMatches
		var allRegularFinals string
		if inclusions == isg.INCLUDEREGULARSEASON {
			allRegularFinals = " AND (round_id = 4 OR round_id = 6) "
		} else if inclusions == isg.INCLUDEFINALSONLY {
			allRegularFinals = " AND (round_id = 2 OR round_id = 3 OR round_id = 1 OR round_id = 5) "
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT match_date from " + sport.TableNameMatches + " where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}
	case "bb":
		if league.LeagueInternalID == 1 {
			// NBA - Use daily_matches tables
			homeSQL = "SELECT home_team_id FROM isg_basketball_daily_matches"
			awaySQL = "SELECT away_team_id FROM isg_basketball_daily_matches "
			var allRegularFinals string
			if inclusions == isg.INCLUDEREGULARSEASON {
				allRegularFinals = " AND season_type_id = 1  "
			} else if inclusions == isg.INCLUDEFINALSONLY {
				allRegularFinals = " AND season_type_id >= 11 AND season_type_id <= 17 "
			}
			if len(seasonIDs) > 0 {
				rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
					" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT match_date from isg_basketball_daily_matches where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
					"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
			}
		} else {
			// NBL - Use round_matches tables
			homeSQL = "SELECT home_team_id FROM isg_basketball_round_matches "
			awaySQL = "SELECT away_team_id FROM isg_basketball_round_matches "
			var allRegularFinals string
			if inclusions == isg.INCLUDEREGULARSEASON {
				allRegularFinals = " AND round_id <= 25 "
			} else if inclusions == isg.INCLUDEFINALSONLY {
				allRegularFinals = " AND round_id > 25 "
			}
			if len(seasonIDs) > 0 {
				rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
					" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT round_id from isg_basketball_round_matches where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
					"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
			}
		}
	case "cr":
		// for big-bash

		homeSQL = "SELECT home_team_id FROM isg_cricket_matches WHERE league_id = " + strconv.Itoa(league.LeagueInternalID)
		awaySQL = "SELECT away_team_id FROM isg_cricket_matches WHERE league_id = " + strconv.Itoa(league.LeagueInternalID)
		var allRegularFinals string
		if league.LeagueInternalID == 1 {
			if inclusions == isg.INCLUDEREGULARSEASON {
				allRegularFinals = " AND round_id >= 1 AND round_id <= 15 "
			} else if inclusions == isg.INCLUDEFINALSONLY {
				allRegularFinals = " AND round_id >= 16 AND round_id <= 17 "
			}
		} else if league.LeagueInternalID == 2 {
			if inclusions == isg.INCLUDEREGULARSEASON {
				allRegularFinals = " AND round_id = 1 "
			} else if inclusions == isg.INCLUDEFINALSONLY {
			} else if league.LeagueInternalID == 2 {
				allRegularFinals = " AND round_id >= 2 AND round_id <= 8 "
			}
		}
		if len(seasonIDs) > 0 {
			rankSQL = "LEFT JOIN isg_ranking_adjustment rank ON rank.sport_id = team.sport_id AND rank.team_id = team.team_id AND rank.league_id = " + strconv.Itoa(league.LeagueInternalID) +
				" AND rank.season_id = " + seasonIDs[len(seasonIDs)-1] + "  AND week_round_date = (SELECT round_id from isg_cricket_matches where status = 'N' AND season_id = " + seasonIDs[len(seasonIDs)-1] + " AND " +
				"league_id = " + strconv.Itoa(league.LeagueInternalID) + " " + allRegularFinals + " ORDER BY Match_date DESC, match_time DESC LIMIT 1)"
		}

	default:
		homeSQL = "SELECT home FROM " + sport.TableNameMatches
		awaySQL = "SELECT away FROM " + sport.TableNameMatches
	}

	// teamSQL = `SELECT entity.entity_api_id, entity.local_id, entity.entity_name FROM isg_api_entities as entity WHERE entity.entity_type="team" AND entity.sport_id = ` + strconv.Itoa(sport.SportInternalID) + ` AND ( local_id IN ( ` + homeSQL + `) OR local_id IN ( ` + awaySQL + `))`
	// rows, err := SportsDb.Query(teamSQL)
	// if err != nil {
	// 	return nil, err
	// }
	customerSQL = " ,filter_name ASC"
	if len(customers) > 0 {
		if customers[0] == 5 || customers[0] == 0 {
			customerSQL = " ,reverse_name ASC"
		}
	}

	switch strings.ToLower(sport.SportID) {
	case "te":
		if len(seasonIDs) == 0 {
			teamSQL = "SELECT distinct IFNULL(isg_api_id,'NULL'), player_id, IFNULL(full_name,'NULL'), IFNULL(filter_name,'NULL'), IFNULL(short_name,'NULL'), " +
				" '' as ranking, player_url_name, flag, '0' AS points, IFNULL(reverse_name,'NULL'), IFNULL(country,'NULL') AS countryName" +
				" FROM isg_tennis_players " +
				" LEFT JOIN isg_country ON isg_country.country_id = isg_tennis_players.country_id " +
				" WHERE level_id =" + strconv.Itoa(league.LeagueInternalID) + " ORDER BY full_name ASC"
		} else if inclusions == 7 && len(seasonNames) == 1 {
			teamSQL = "SELECT distinct IFNULL(isg_api_id,'NULL'), player.player_id, IFNULL(full_name,'NULL'), IFNULL(filter_name,'NULL'), IFNULL(short_name,'NULL'), " +
				" IFNULL(ranking,9999) AS ranking, IFNULL(player.player_url_name,'NULL'), IFNULL(country.flag,'NULL'),IFNULL(player_season.points,0), IFNULL(reverse_name,'NULL'),IFNULL(country,'NULL') AS countryName " +
				" FROM  isg_tennis_players player " +
				" LEFT JOIN isg_country country ON country.country_id = player.country_id " +
				" LEFT JOIN isg_tennis_player_season AS player_season ON player_season.player_id = player.player_id AND FIND_IN_SET(player_season.season_id, '" + result + "')" +
				" where   player.level_id = " + strconv.Itoa(league.LeagueInternalID) +
				" ORDER BY ranking ASC " + customerSQL
		} else {
			teamSQL = "SELECT distinct IFNULL(isg_api_id,'NULL'), tmp.player_id, IFNULL(full_name,'NULL'), IFNULL(filter_name,'NULL'), IFNULL(short_name,'NULL'), " +
				" IFNULL(ranking,9999) AS ranking, b.player_url_name, country.flag,IFNULL(c.points,0), IFNULL(reverse_name,'NULL'),IFNULL(country.country,'NULL') AS countryName " +
				" FROM (" +
				" SELECT distinct player1_id AS player_id" +
				" FROM isg_tennis_matches" +
				" WHERE level_id =" + strconv.Itoa(league.LeagueInternalID) + " AND FIND_IN_SET(season_id, '" + result + "')  " +
				" UNION ALL" +
				" SELECT distinct player2_id AS player_id" +
				" FROM isg_tennis_matches" +
				" WHERE level_id =" + strconv.Itoa(league.LeagueInternalID) + " AND FIND_IN_SET(season_id, '" + result + "')) AS tmp " +
				" LEFT JOIN isg_tennis_players AS b ON b.player_id = tmp.player_id" +
				" LEFT JOIN isg_country country ON country.country_id = b.country_id " +
				" LEFT JOIN isg_tennis_player_season AS c ON c.player_id = b.player_id AND c.season_id = " + seasonIDs[len(seasonIDs)-1] + " ORDER BY ranking ASC" + customerSQL

		}

	default:
		if rankSQL != "" {
			teamSQL = "SELECT DISTINCT team.isg_api_id, team.team_id, team.team_name, team.short_teamname, team.pinnaclesports, team.unibet, IFNULL(team.abbreviation,''),team.sort_val,team.url,team.icon,IFNULL(rank.ranking,'9999') " +
				" FROM isg_team team " + rankSQL +
				" WHERE team.sport_id = " + strconv.Itoa(sport.SportInternalID) +
				" AND ( team.team_id IN ( " + homeSQL + ") OR team.team_id IN ( " + awaySQL + ") ) ORDER BY rank.ranking ASC "
		} else {
			teamSQL = "SELECT DISTINCT isg_api_id, team_id,team_name, short_teamname, pinnaclesports, unibet, abbreviation,sort_val,url,icon" +
				" FROM isg_team team" +
				" WHERE team.sport_id = " + strconv.Itoa(sport.SportInternalID) +
				" AND team_id IN ( " + homeSQL + ") ORDER BY team_name "
		}
		//fmt.Println(teamSQL)
	}

	//teamSQL = `SELECT entity.entity_api_id, entity.local_id, entity.entity_name FROM isg_api_entities as entity WHERE entity.entity_type="team" AND entity.sport_id = ` + strconv.Itoa(sport.SportInternalID) + ` AND ( local_id IN ( ` + homeSQL + `) OR local_id IN ( ` + awaySQL + `))`
	rows, err := SportsDb.Query(teamSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var team isg.Team
		switch strings.ToLower(sport.SportID) {
		case "te":
			err := rows.Scan(&team.TeamID, &team.TeamInternalID, &team.TeamName, &team.TeamShortName, &team.TeamAltName1, &team.Ranking, &team.TeamURL, &team.TeamFlag, &team.Points, &team.TeamReverseName, &team.Country)
			if err != nil {
				fmt.Println(err)
				fmt.Println(teamSQL)
				fmt.Println(team)
				return nil, err
			}
		default:
			if rankSQL != "" {
				err := rows.Scan(&team.TeamID, &team.TeamInternalID, &team.TeamName, &team.TeamShortName, &team.TeamAltName1, &team.TeamAltName2, &team.Abbreviation, &team.SortVal, &team.TeamURL, &team.TeamFlag, &team.Ranking)

				if err != nil {
					fmt.Println(err)
					fmt.Println(teamSQL)
					fmt.Println(team)
					return nil, err
				}
			} else {
				err := rows.Scan(&team.TeamID, &team.TeamInternalID, &team.TeamName, &team.TeamShortName, &team.TeamAltName1, &team.TeamAltName2, &team.Abbreviation, &team.SortVal, &team.TeamURL, &team.TeamFlag)
				if err != nil {
					fmt.Println(err)
					fmt.Println(teamSQL)
					fmt.Println(team)
					return nil, err
				}
			}

		}

		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil

}

// GetSport returns an object containing essential info of a sport.
// The sport input can be a name or an Id.
func GetSport(sport string) (isg.Sport, error) {
	var tblmatches, tblplayers, tblseasons sql.NullString
	var result isg.Sport
	var query string

	if _, err := strconv.Atoi(sport); err == nil {

		// Check if cached data exists
		sportIntId, err := strconv.Atoi(sport)
		if err == nil {
			id := SportAPIIDs[sportIntId]
			result, valid := SportObjects[id]
			if valid {
				return result, nil
			}
			// cacheData, err := CacheGet("isgapi:obj:sport:" + id)
			// if err == nil && cacheData != nil {
			// 	returnData := isg.Extract(cacheData)
			// 	err = json.Unmarshal(returnData, &result)
			// 	return result, nil
			// }
		}

		// sport value is numeric, most likely an ID
		query = "SELECT sport_id as id, sport_name as name, sport_api_code as apicode, sport_match_tablename as matchtable, " +
			" sport_player_tablename as playertable, sport_season_tablename as seasontable, sport_url FROM isg_sports WHERE sport_id = ?"
		err = SportsDb.QueryRow(query, sport).Scan(
			&result.SportInternalID,
			&result.SportName,
			&result.SportID,
			&tblmatches,
			&tblplayers,
			&tblseasons,
			&result.SportURL,
		)
		if err != nil {
			return result, err
		}

	} else {

		id, valid := ValidSportIDs[sport]
		if valid {
			result, valid := SportObjects[id]
			if valid {
				return result, nil
			}
			// cacheData, err := CacheGet("isgapi:obj:sport:" + id)
			// if err == nil && cacheData != nil {
			// 	returnData := isg.Extract(cacheData)
			// 	err = json.Unmarshal(returnData, &result)
			// 	return result, nil
			// }
		}

		id, valid = SportIDForName[sport]
		if valid {
			result, valid := SportObjects[id]
			if valid {
				return result, nil
			}
		}

		query = "SELECT sport_id as id, sport_name as name, sport_api_code as apicode, sport_match_tablename as matchtable, " +
			" sport_player_tablename as playertable, sport_season_tablename as seasontable, sport_url FROM isg_sports WHERE sport_name = ?  or sport_api_code = ? or sport_api_altname = ? or sport_url = ? "
		err = SportsDb.QueryRow(query, sport, sport, sport, sport).Scan(
			&result.SportInternalID,
			&result.SportName,
			&result.SportID,
			&tblmatches,
			&tblplayers,
			&tblseasons,
			&result.SportURL,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return result, errors.New("invalid sport")
			}
			return result, err
		}

	}

	if tblmatches.Valid {
		result.TableNameMatches = tblmatches.String
	}
	if tblplayers.Valid {
		result.TableNamePlayers = tblplayers.String
	}
	if tblseasons.Valid {
		result.TableNameSeasons = tblseasons.String
	}

	// Cache it for future use
	resultBytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println(err)
	} else {
		saveData := isg.Compress(resultBytes)
		err = CacheSet("isgapi:obj:sport:"+CleanText(result.SportID, true, true), saveData)
		if err != nil {
			fmt.Println("Error has occured while converting the data")
		}
	}

	return result, nil
}

// GetSportLeague :
func GetSportLeague(sportid int, leaguename string) (int, string, error) {

	var sqlstr, league string
	var leagueid int
	switch sportid {

	case 1: // Aussie Rules
		sqlstr = "SELECT league_id,league_name FROM isg_aussie_rules_leagues WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "') "
	case 2: //American Football
		sqlstr = "SELECT league_id, first_name FROM isg_american_football_league WHERE status = ? AND (league_url = '" + leaguename + "' OR first_name = '" + leaguename + "') "

	case 3: //Basketball
		sqlstr = "SELECT league_id, league_name FROM isg_basketball_leagues WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "') "

	case 4: // Soccer
		sqlstr = "SELECT league_id, first_name FROM isg_soccer_leagues WHERE (status = ? OR status = 0) AND (league_url = '" + leaguename + "' OR first_name = '" + leaguename + "')"

	case 6: // Tennis
		sqlstr = "SELECT level_id,level_name FROM isg_tennis_tournament_level WHERE status = ? AND (level_url = '" + leaguename + "' OR level_name = '" + leaguename + "')"

	case 7: // Rugby League
		sqlstr = "SELECT league_id, league_name FROM isg_rugby_league WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "')"

	case 8: // Hockey
		sqlstr = "SELECT league_id,league_name FROM isg_hockey_leagues WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "')"

	case 9: // Baseball
		sqlstr = "SELECT league_id,league_name FROM isg_baseball_leagues WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "')"

	case 10: // Rugby Union
		sqlstr = "SELECT league_id, league_name FROM isg_rugby_union_league WHERE status = ? AND (league_url = '" + leaguename + "' OR league_name = '" + leaguename + "')"
	}

	err := SportsDb.QueryRow(sqlstr, 1).Scan(&leagueid, &league)

	return leagueid, league, err

}

// GetTeam :
func GetTeam(sportid int, teamName string) (int, error) {
	var teamid int
	err := SportsDb.QueryRow("select team_id from isg_team where sport_id = ? AND ((team_name = ?) OR (isg_api_name = ?) OR (isg_api_regionname = ?) OR (filtername = ?) OR (url = ?))", sportid, teamName, teamName, teamName, teamName, teamName).Scan(&teamid)
	if err != nil {
		return 0, err
	}
	return teamid, nil
}

// GetTennisPlayer :
func GetTennisPlayer(levelid int, playerName string) (int, error) {
	var playerid int
	err := SportsDb.QueryRow("select player_id from isg_tennis_players where level_id = ? AND ((full_name = ?) OR (player_url_name = ?) OR (isg_api_id = ?))", levelid, playerName, playerName, playerName).Scan(&playerid)
	if err != nil {
		return 0, err
	}
	return playerid, nil
}

// GetAFLRoundTeamScore :
func GetAFLRoundTeamScore(seasons []int, typefunction string, curseasonid, curround int) []isg.AllMatchDetail {
	var matches []isg.AllMatchDetail
	var seasonid string
	for i := 0; i < len(seasons); i++ {
		if seasonid == "" {
			seasonid = seasonid + strconv.Itoa(seasons[i])
		} else {
			seasonid = seasonid + "," + strconv.Itoa(seasons[i])
		}
	}
	sql := `SELECT round_id, season_id, ` + typefunction + `(scores.home_score) AS maxhome, ` + typefunction + `(scores.away_score) AS maxaway` +
		` FROM isg_aussie_rules_matches AS matches ` +
		` LEFT JOIN isg_aussie_rules_matches_scores scores ON scores.match_id = matches.match_id` +
		` WHERE FIND_IN_SET(season_id, ?) round_id <= 24 and matches.status='N' GROUP BY round_id, season_id ORDER BY match_date ASC`
	fmt.Println(sql)
	rows, err := SportsDb.Query(sql, seasonid)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		var match isg.AllMatchDetail
		var round, season int
		var max, maxhome, maxaway float64
		rows.Scan(&round, &season, &maxhome, &maxaway)
		max = math.Max(maxhome, maxaway)
		if typefunction == "min" {
			max = math.Min(maxhome, maxaway)
		}

		if season == curseasonid && round > curround {
			continue
		}

		sql := `Select home_score, away_score FROM isg_aussie_rules_matches AS matches` +
			` LEFT JOIN isg_aussie_rules_matches_scores scores ON scores.match_id = matches.match_id` +
			` WHERE round_id = ? and season_id = ? having (home_score = ? OR away_score = ?) `

		rowss, err := SportsDb.Query(sql, round, season, max, max)
		if err != nil {
			fmt.Println(err)
		}
		defer rowss.Close()
		rowData := StringArray(rowss)
		match.MaxScore = int(max)
		match.RoundID = round
		match.Count = len(rowData)
		match.Season = season
		matches = append(matches, match)
	}
	return matches
}

// GetAFLRoundHighestScoreOrMargin :
func GetAFLRoundHighestScoreOrMargin(season []int, typefunction, sign string, curseasonid, curround int) []isg.AllMatchDetail {
	var matches []isg.AllMatchDetail
	var seasonid string
	for i := 0; i < len(season); i++ {
		if seasonid == "" {
			seasonid = seasonid + strconv.Itoa(season[i])
		} else {
			seasonid = seasonid + "," + strconv.Itoa(season[i])
		}
	}
	sql := `Select ` + typefunction + `(ABS(home_score ` + sign + ` away_score)) AS maxscore, round_id , season_id FROM isg_aussie_rules_matches AS matches` +
		`  LEFT JOIN isg_aussie_rules_matches_scores scores ON scores.match_id = matches.match_id` +
		` WHERE FIND_IN_SET(season_id, ?) and round_id <= 24 and matches.status='N' group by round_id, season_id  ORDER BY match_date asc  `

	rows, err := SportsDb.Query(sql, seasonid)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var match isg.AllMatchDetail
		var round, score, seasons int
		rows.Scan(&score, &round, &seasons)
		if seasons == curseasonid && round > curround {
			continue
		}
		sql := `Select home_score, away_score FROM isg_aussie_rules_matches AS matches` +
			` LEFT JOIN isg_aussie_rules_matches_scores scores ON scores.match_id = matches.match_id` +
			` WHERE round_id = ? and season_id = ? having (ABS(scores.home_score ` + sign + ` scores.away_score)) = ? `
		rowss, err := SportsDb.Query(sql, round, seasons, score)
		if err != nil {
			fmt.Println(err)
		}
		defer rowss.Close()
		rowData := StringArray(rowss)

		match.MaxScore = score
		match.RoundID = round
		match.Count = len(rowData)
		match.Season = seasons
		matches = append(matches, match)
	}
	return matches
}

func GetEntitiesInSeason(sport, season string) ([]isg.PortalEntity, error) {
	results := []isg.PortalEntity{}

	// Find cached list if available

	// If no cache found, query database and fill in PortalEntity array

	return results, nil
}

// VerifySeasonValues checks the season values string (+ delimited e.g. 2015+2016+2017) and makes sure
// all season values are correct for the specified sport.
func VerifySeasonValues(sport isg.Sport, league isg.League, seasonValues string) (bool, string, error) {
	validSeasons, err := AllSeasonsBySport(sport, league)
	if err != nil {
		return false, "", err
	}

	seasonList := strings.Split(seasonValues, "+")
	for _, season := range seasonList {
		isValid := false
		for _, validseason := range validSeasons {
			if validseason.SeasonID == strings.TrimSpace(season) {
				isValid = true
			}
		}
		if !isValid {
			return false, season, nil
		}
	}

	return true, "", nil
}

// AllSeasonsBySport returns the list of seasons for a specific sport and league.
func AllSeasonsBySport(sport isg.Sport, league isg.League) ([]isg.Season, error) {
	seasons := []isg.Season{}
	var sqlquery, seasonTable string
	switch sport.SportInternalID {
	case 1: // AFL
		return []isg.Season{}, nil
	case 2: // NFL
		sqlquery = "select season_id, season from " + sport.TableNameSeasons + " where status = 1"
	case 3: // NBA
		sqlquery = "select season_id, season_url as season from " + sport.TableNameSeasons + " where status = 1"
	case 4: // Soccer
		seasonTable = "isg_soccer_season"
		if league.LeagueInternalID == 17 || league.LeagueInternalID == 23 {
			seasonTable = "isg_soccer_worldcup_season"
		}
		if league.LeagueInternalID == 19 || league.LeagueInternalID == 24 || league.LeagueInternalID == 25 || league.LeagueInternalID == 26 ||
			league.LeagueInternalID == 27 || league.LeagueInternalID == 28 || league.LeagueInternalID == 29 {
			sport.TableNameSeasons = "isg_sports_league_seasons"
			seasonTable = "isg_sports_league_seasons"
		}

		if league.LeagueInternalID == 19 || league.LeagueInternalID == 24 || league.LeagueInternalID == 25 || league.LeagueInternalID == 26 ||
			league.LeagueInternalID == 27 || league.LeagueInternalID == 28 || league.LeagueInternalID == 29 {
			sqlquery = "select season_id, season_url from " + seasonTable + " where status = 1 "
		} else if league.LeagueInternalID == 23 {
			sqlquery = "select season_id, isg_api_id from " + seasonTable + " where (status = 1 OR status = 0)"
		} else {
			sqlquery = "select season_id, isg_api_id from " + seasonTable + " where status = 1"
		}

	case 5: // Cricket
		seasonTable = "isg_cricket_season"
		if league.LeagueInternalID == 2 {
			seasonTable = "isg_cricket_single_season"
		}
		sqlquery = "select season_id, season_url from " + seasonTable + " where status = 1"
	case 6: // Tennis
		sqlquery = "select season_id, season from " + sport.TableNameSeasons + " where status = 1"
	case 7: // NRL
		sqlquery = "select season_id, season from " + sport.TableNameSeasons + " where status = 1"
	case 8: // Ice Hockey
		sqlquery = "select season_id, season_url from " + sport.TableNameSeasons + " where status = 1"
	case 9: // MLB
		sqlquery = "select season_id, season from " + sport.TableNameSeasons + " where status = 1"
	case 10: // Super Rugby
		sqlquery = "select season_id, season from " + sport.TableNameSeasons + " where status = 1"
	}

	// err := SportsDb.QueryRow(sql, season).Scan(&seasonid)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return 0, err
	// }

	rows, err := SportsDb.Query(sqlquery)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		s := isg.Season{sport.SportID, sport.SportInternalID, league.LeagueID, league.LeagueInternalID, "", 0, "", ""}
		err := rows.Scan(&s.SeasonInternalID, &s.SeasonID)
		//	fmt.Println(s.SeasonInternalID, s.SeasonID)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}
	return seasons, nil

}

// GetSeasonID returns the proper season Id based on a sport and season name.
func GetSeasonID(sport isg.Sport, season string) (int, error) {
	var seasonid int
	var sql string
	switch sport.SportInternalID {
	case 1: // AFL
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	case 2: // NFL
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	case 3: // NBA
		sql = "select season_id from " + sport.TableNameSeasons + " where season_url = ?"
	case 4: // Soccer
		if sport.TableNameSeasons == "isg_sports_league_seasons" {
			sql = "select season_id from " + sport.TableNameSeasons + " where season_url = ?"
		} else {
			sql = "select season_id from " + sport.TableNameSeasons + " where isg_api_id = ?"
		}

	case 5: // Cricket
		sql = "select season_id from " + sport.TableNameSeasons + " where season_url = ?"
	case 6: // Tennis
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	case 7: // NRL
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	case 8: // Ice Hockey
		sql = "select season_id from " + sport.TableNameSeasons + " where season_url = ?"
	case 9: // MLB
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	case 10: // Super Rugby
		sql = "select season_id from " + sport.TableNameSeasons + " where season = ?"
	}
	err := SportsDb.QueryRow(sql, season).Scan(&seasonid)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return seasonid, nil

}

func GetLeagueID(sport isg.Sport, league string) (isg.League, error) {
	id := strconv.Itoa(sport.SportInternalID)
	//fmt.Println(SportsLeagues[id])
	sportLeagues := SportsLeagues[id]

	for _, s := range sportLeagues {
		if strings.ToLower(league) == s.LeagueID {
			return s, nil
		}
		if strings.ToLower(league) == s.LeagueEntityKey {
			return s, nil
		}
	}

	return isg.League{}, errors.New("No matching league found")
}

// GetNRLMaxRounds : get the max round
func GetNRLMaxRounds() ([]isg.NRLRounds, error) {
	var results []isg.NRLRounds
	rows, err := SportsDb.Query("SELECT season_id as seasonid,max(round_id) as maxround FROM isg_rugby_league_matches WHERE status='N' AND round_id<=26 GROUP BY season_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nrlround isg.NRLRounds
		err := rows.Scan(&nrlround.SeasonID, &nrlround.MaxRound)
		if err != nil {
			return nil, err
		}
		results = append(results, nrlround)
	}

	return results, nil
}

// GetMarkets returns all markets for a sport/league.
func GetMarkets(sport isg.Sport, league isg.League) ([]isg.Market, error) {
	var results []isg.Market

	rows, err := SportsDb.Query("select isg_api_id, market_name, full_name from isg_market where sport_id = ? AND league_level_id = ? ", sport.SportInternalID, league.LeagueInternalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var market isg.Market
		err := rows.Scan(&market.MarketID, &market.MarketShortName, &market.MarketName)
		if err != nil {
			return nil, err
		}
		results = append(results, market)
	}

	return results, nil
}

// GetMarketById returns the market object for a specific market
func GetMarketById(sport isg.Sport, league isg.League, marketID string) (isg.Market, error) {

	rows, err := SportsDb.Query("select isg_api_id, market_name, full_name from isg_market where sport_id = ? AND league_level_id = ? AND isg_api_id = ?", sport.SportInternalID, league.LeagueInternalID, marketID)
	if err != nil {
		return isg.Market{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var market isg.Market
		err := rows.Scan(&market.MarketID, &market.MarketShortName, &market.MarketName)
		if err != nil {
			return isg.Market{}, err
		}
		return market, nil
	}

	return isg.Market{}, errors.New("Market not found")
}

// GetPointsAdjustments returns all the points adjustments for a particular team.
func GetPointsAdjustments(sport isg.Sport, league isg.League, team isg.Team) ([]isg.PointsAdjustments, error) {
	var results []isg.PointsAdjustments

	rows, err := SportsDb.Query("select sport_id, season_id, team_id, adjustment from isg_points_adjustment where sport_id = ? AND league_id = ? AND team_id = ?", sport.SportInternalID, league.LeagueInternalID, team.TeamInternalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var adjustment isg.PointsAdjustments
		err := rows.Scan(&adjustment.SportID, &adjustment.SeasonID, &adjustment.TeamID, &adjustment.Points)
		if err != nil {
			return nil, err
		}
		results = append(results, adjustment)
	}

	return results, nil
}

func CleanText(text string, lowerCase bool, trimDoubleSpaces bool) string {
	sanitizedText := sanitize.HTML(text)
	sanitizedText = strings.TrimSpace(sanitizedText)
	sanitizedText = strings.Replace(sanitizedText, "\t", " ", -1)
	if lowerCase {
		sanitizedText = strings.ToLower(sanitizedText)
	}
	if trimDoubleSpaces {
		for {
			if strings.Contains(sanitizedText, "  ") {
				sanitizedText = strings.Replace(sanitizedText, "  ", " ", -1)
			} else {
				break
			}
		}
	}

	return sanitizedText
}

// GetCurrentRound :
func GetCurrentAFLRound() (int, int, error) {
	var round, seasonid int
	query := "SELECT round_id, season_id FROM isg_aussie_rules_matches WHERE status = ? LIMIT 1"
	err := SportsDb.QueryRow(query, `Y`).Scan(&round, &seasonid)
	if err != nil {
		query := "SELECT round_id, season_id FROM isg_aussie_rules_matches WHERE match_date <= CURDATE() ORDER BY match_date DESC LIMIT 1"
		err2 := SportsDb.QueryRow(query).Scan(&round, &seasonid)
		if err2 != nil {
			return 0, 0, errors.New("Current AFL round not found")
		}
		return seasonid, round, nil
	}
	return seasonid, (round - 1), nil
}

// GetProvider :
func GetProvider(providerName string) (int, error) {
	var providerid int

	if CleanText(providerName, true, true) == "draftkings" {
		providerName = "unibet"
	}
	err := SportsDb.QueryRow("select provider_id from isg_providers WHERE ((provider_name = ?) OR (provider_url = ?))", providerName, providerName).Scan(&providerid)
	if err != nil {
		return 0, err
	}
	return providerid, nil
}

// GetCurrentRoundandCount - Get current round and count of y matches in that round for rugby union
func GetCurrentRoundandCount() ([]isg.Ycount, error) {
	var round, seasonid, count int
	var ycount []isg.Ycount
	query := "SELECT round_id, season_id FROM isg_rugby_union_matches WHERE status = ? LIMIT 1"
	err := SportsDb.QueryRow(query, `Y`).Scan(&round, &seasonid)
	if err != nil {
		return ycount, err
	}
	query2 := "SELECT count(status= 'Y') FROM isg_rugby_union_matches WHERE round_id = ? AND season_id = ?"
	err = SportsDb.QueryRow(query2, round, seasonid).Scan(&count)
	var result isg.Ycount
	result.SeasonID = seasonid
	result.RoundID = round
	result.Count = count
	ycount = append(ycount, result)
	return ycount, nil
}

// GetMatchInfo :
func GetMatchInfo(objsport isg.Sport, leagueid, matchid int) (isg.ErrorLogRecord, error) {
	var sqlstr string
	var result isg.ErrorLogRecord

	switch objsport.SportInternalID {
	case 1: // Aussie Rules
		sqlstr = "SELECT match_id, season_id, league_id, round_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	case 2: // American Football
		sqlstr = "SELECT match_id, season_id, 0 as league_id, week_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? "

	case 3: // Basketball
		if leagueid == 1 {
			sqlstr = "SELECT match_id, season_id, league_id, season_type_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
				" FROM " + objsport.TableNameMatches + " AS matches " +
				" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
				" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)
		} else {

			sqlstr = "SELECT match_id, season_id, league_id, round_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
				" FROM " + objsport.TableNameMatches + " AS matches " +
				" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
				" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)
		}

	case 4: // Soccer
		sqlstr = "SELECT match_id, season_id, league_id, week_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	case 6: // Tennis
		sqlstr = "SELECT match_id, season_id, matches.level_id, tournament_id, match_date, match_time, h_team.full_name as hometeam, a_team.full_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_tennis_players AS h_team ON h_team.player_id = matches.player1_id " +
			" LEFT JOIN isg_tennis_players AS a_team ON a_team.player_id = matches.player2_id " +
			" WHERE match_id = ? AND matches.level_id = " + strconv.Itoa(leagueid)

	case 7: // Rugby League
		sqlstr = "SELECT match_id, season_id, league_id, round_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	case 8: // Hockey
		sqlstr = "SELECT match_id, season_id, league_id, week_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	case 9: // Baseball
		sqlstr = "SELECT match_id, season_id, league_id, round_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	case 10: // Rugby Union
		sqlstr = "SELECT match_id, season_id, league_id, round_id, match_date, match_time, h_team.team_name as hometeam, a_team.team_name as awayteam " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" LEFT JOIN isg_team AS h_team ON h_team.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS a_team ON a_team.team_id = matches.away_team_id " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	}

	err := SportsDb.QueryRow(sqlstr, matchid).Scan(
		&result.MatchID,
		&result.SeasonID,
		&result.LeagueID,
		&result.RoundWeek,
		&result.MatchDate,
		&result.MatchTime,
		&result.HomeTeam,
		&result.AwayTeam,
	)
	result.ProviderName = "isports"
	result.ScriptName = "Genius Odds"
	if err == sql.ErrNoRows {
		return result, nil
	} else if err != nil {
		return result, err
	}
	return result, nil

}

//InsertErrorLog : Insert All Error log record according to cron, error type etc.
func InsertErrorLog(objerror isg.ErrorLogRecord, sportid, alertid int, tablename, errormsg string) {

	stmt, err := LogDb.Prepare("INSERT INTO " + tablename + " SET alert_id = ?, alert_group_id = ?, sport_id = ?, match_id = ?, season_id = ?, round_week = ?, league_id = ?, home_team = ?, away_team = ?, match_date = ?, match_time = ?, script_name = ?, error_msg = ?, dateadded = ?, provider_name = ? ON DUPLICATE KEY UPDATE error_msg = ?, dateadded = ? ")
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = stmt.Exec(alertid, sportid, sportid, objerror.MatchID, objerror.SeasonID, objerror.RoundWeek, objerror.LeagueID, objerror.HomeTeam, objerror.AwayTeam, objerror.MatchDate, objerror.MatchTime, objerror.ScriptName, errormsg, time.Now().In(AEST).Format("2006-01-02 15:04:05"), objerror.ProviderName, errormsg, time.Now().In(AEST).Format("2006-01-02 15:04:05"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

// GetMatchIDFromEventID : get the matchid from the passed eventid
func GetMatchIDFromEventID(providerid int, eventid string, sportID int) (int, int, int, error) {

	if providerid == 15 {
		providerid = 5
	}
	var matchid, sportid, leagueid int
	var err error
	if sportID != 0 {
		err = SportsDb.QueryRow("SELECT match_id,sport_id, league_id FROM tblform_matcheventsmapping WHERE customer_id=? AND event_id=? and sport_id = ? AND enabled = ? ORDER BY last_update DESC LIMIT 0,1", providerid, eventid, sportID, 1).Scan(&matchid, &sportid, &leagueid)
	} else {
		err = SportsDb.QueryRow("SELECT match_id,sport_id, league_id FROM tblform_matcheventsmapping WHERE customer_id=? AND event_id=? AND enabled = ? ORDER BY last_update DESC LIMIT 0,1", providerid, eventid, 1).Scan(&matchid, &sportid, &leagueid)
	}

	if err == sql.ErrNoRows {
		return 0, 0, 0, errors.New("event not found")
	}
	if err != nil {
		return 0, 0, 0, err
	}
	return matchid, sportid, leagueid, nil
}

// GetLeagueIsEnable : get the league is active or not for customer wise
func GetLeagueIsEnable(customerID, sportID, leagueID int) (int, error) {
	var cnt int
	err := SportsDb.QueryRow("SELECT COUNT(1) AS cnt FROM tblform_matcheventsmapping_exclusions WHERE customer_id = ? AND sport_id = ? AND league_id = ? ",
		customerID, sportID, leagueID).Scan(&cnt)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return cnt, nil
}

// GetMatchStatus : get the current status of the match
func GetMatchStatus(sportid, matchid, leagueid int) (string, error) {
	var sqlstr, status string
	switch sportid {
	case 1: // Aussie Rules
		sqlstr = "SELECT status FROM isg_aussie_rules_matches WHERE match_id = ?"

	case 2: // American Football
		sqlstr = "SELECT status FROM isg_nflmatches WHERE match_id = ?"

	case 3: // Basketball
		if leagueid == 1 {
			sqlstr = "SELECT status FROM isg_basketball_daily_matches WHERE match_id = ?"
		} else {
			sqlstr = "SELECT status FROM isg_basketball_round_matches WHERE match_id = ?"
		}

	case 4: // Soccer
		sqlstr = "SELECT status FROM isg_soccermatches WHERE match_id=? "

	case 6: // Tennis
		sqlstr = "SELECT status FROM isg_tennis_matches WHERE match_id = ?"

	case 7: // Rugby League
		sqlstr = "SELECT status FROM isg_rugby_league_matches WHERE match_id=?"

	case 8: // Ice Hockey
		sqlstr = "SELECT status FROM isg_hockeymatches WHERE match_id = ?"

	case 9: // Baseball
		sqlstr = "SELECT status FROM isg_baseball_matches WHERE match_id = ?"

	case 10: // Rugby Union
		sqlstr = "SELECT status FROM isg_rugby_league_matches WHERE match_id=?"

	}
	err := SportsDb.QueryRow(sqlstr, matchid).Scan(&status)
	if err == sql.ErrNoRows {
		return status, errors.New("match not found")
	} else if err != nil {
		return status, err
	}
	return status, nil
}

//----------------------------------------------------
// SportDB AU database related queries
//---------------------------------------------------
// GetAuroraTeam :
func GetAuroraTeam(sportid int, teamName string) (int, error) {
	var teamid int
	err := SportsDbAU.QueryRow("select team_id from isg_team_v2 where sport_id = ? AND ((team_name = ?)  OR (filter_name = ?) OR (team_url = ?))", sportid, teamName, teamName, teamName).Scan(&teamid)
	if err != nil {
		return 0, err
	}
	return teamid, nil
}

//GetAussieRulesRoundID used for get round id from round table
func GetAussieRulesRoundID(round string) (int, error) {
	var roundID int
	err := SportsDbAU.QueryRow("SELECT round_id from isg_aussie_rules_round_v2 where round_name = ? OR round_url = ? OR sb_round_name = ?", round, round, round).Scan(&roundID)
	if err == sql.ErrNoRows {
		return 0, errors.New("round not found")
	}
	if err != nil {
		return 0, err
	}
	return roundID, nil
}

//GetAussieRulesSeasonID used for get  season id from season table
func GetAussieRulesSeasonID(season string) (int, error) {
	var seasonID int
	err := SportsDbAU.QueryRow("SELECT season_id from isg_aussie_rules_season_v2 where season = ?", season).Scan(&seasonID)
	if err == sql.ErrNoRows {
		return 0, errors.New("round not found")
	}
	if err != nil {
		return 0, err
	}
	return seasonID, nil
}

//GetAussieRulesMatchID used for getting match id as per hometeamid / awyteamid / roundid / seasonid
func GetAussieRulesMatchID(hometeamid, awayteamid, roundid, seasonid int) (int, error) {
	var matchid int
	sqlstr := "SELECT match_id from isg_aussie_rules_matches_v2 where home_team_id = ? and away_team_id = ? and round_id = ? and season_id = ?"
	err := SportsDbAU.QueryRow(sqlstr, hometeamid, awayteamid, roundid, seasonid).Scan(&matchid)

	if err == sql.ErrNoRows {
		return 0, errors.New("matchid  not found")
	}
	if err != nil {
		return 0, err
	}
	return matchid, nil
}

// GetSportsSeasonList :
func GetSportsSeasonList(objsport isg.Sport, seasons string) []isg.Season {

	var sqlstr, sqlWhere string
	seasonIDs := []isg.Season{}

	if seasons != "all" {
		seasonArr := strings.Replace(seasons, "+", "','", -1)
		sqlWhere = " AND (season_url IN ('" + seasonArr + "') OR season IN ('" + seasonArr + "')) "
	}

	switch objsport.SportID {

	case "ar":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 7 " + sqlWhere + " ORDER BY season_id DESC"

	case "af":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 19 " + sqlWhere + " ORDER BY season_id DESC"

	case "bb":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "sc":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 5 " + sqlWhere + " ORDER BY season_id DESC"

	case "cr":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "te":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "rl":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "bl":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE  season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "ih":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	case "ru":
		sqlstr = "SELECT season_id, season, season_url FROM " + objsport.TableNameSeasons + " WHERE season_id > 0 " + sqlWhere + " ORDER BY season_id DESC"

	}

	rows, err := SportsDb.Query(sqlstr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		record := isg.Season{}
		err := rows.Scan(&record.SeasonID, &record.SeasonName, &record.SeasonURL)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		seasonIDs = append(seasonIDs, record)
	}
	if len(strings.Split(seasons, "+")) != len(seasonIDs) && seasons != "all" {
		return nil
	}
	return seasonIDs
}

// GetRoundWeekDetails :
func GetRoundWeekDetails(objsport isg.Sport, objleague isg.League, roundstr string) ([]isg.SportRound, string, error) {

	var objSprotRounds []isg.SportRound
	var sqlstr, sportweekround string
	//var weekround int
	switch objsport.SportID {

	case "ar":
		sportweekround = "round"
		sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_aussie_rules_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"

	case "af":
		sportweekround = "week"
		sqlstr = "SELECT week_id, week_name, short_week_name, week_url FROM isg_nfl_week WHERE (short_week_name IN ('" + roundstr + "') OR week_url IN ('" + roundstr + "'))"

	case "bb":
		if objleague.LeagueInternalID == 1 {
			sportweekround = "seasontype"
			sqlstr = "SELECT season_type_id, season_type_name, short_type_name, type_url FROM isg_basketball_season_type WHERE (short_type_name IN ('" + roundstr + "') OR type_url IN ('" + roundstr + "'))"
		} else {
			sportweekround = "round"
			sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_basketball_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"
		}

	case "sc":

		if objleague.LeagueInternalID == 10 || objleague.LeagueInternalID == 11 {
			sportweekround = "matchday"
			sqlstr = "SELECT match_day_id, match_day, short_match_day_name, match_day_url FROM isg_soccer_match_days WHERE (short_match_day_name IN ('" + roundstr + "') OR match_day_url IN ('" + roundstr + "') OR match_day_group_url IN ('" + roundstr + "'))"
		} else if objleague.LeagueInternalID == 17 {
			sportweekround = "week"
			sqlstr = "SELECT week_id, week, short_week_name, week_url FROM isg_soccer_worldcup_week WHERE (short_week_name IN ('" + roundstr + "')  OR week_url IN ('" + roundstr + "'))"
		} else {
			sportweekround = "week"
			sqlstr = "SELECT week_id, week, short_week_name, week_url FROM isg_soccer_week WHERE (short_week_name IN ('" + roundstr + "') OR week_url IN ('" + roundstr + "'))"
		}

	case "cr":
		if objleague.LeagueInternalID == 1 {
			sportweekround = "round"
			sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_cricket_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"
		} else {
			sportweekround = "seasontype"
			sqlstr = "SELECT season_type_id, season_type_name, short_type_name, type_url FROM isg_cricket_season_type WHERE (short_type_name IN ('" + roundstr + "') OR type_url IN ('" + roundstr + "'))"
		}

	case "te":
		sportweekround = "round"
		sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_tennis_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"

	case "rl":
		sportweekround = "round"
		sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_rugby_league_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"

	case "ih":
		sportweekround = "week"
		sqlstr = "SELECT week_id, week_name, short_week_name, week_url FROM isg_hockey_week WHERE (short_week_name IN ('" + roundstr + "') OR week_url IN ('" + roundstr + "'))"

	case "bl":
		sportweekround = "round"
		sqlstr = "SELECT round_id, round_name, round_short_name, round_url FROM isg_baseball_round WHERE (round_short_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"

	case "ru":
		sportweekround = "round"
		sqlstr = "SELECT round_id, round_name, short_round_name, round_url FROM isg_rugby_union_round WHERE (short_round_name IN ('" + roundstr + "') OR round_url IN ('" + roundstr + "'))"
	}

	rows, err := SportsDb.Query(sqlstr)
	if err != nil {
		return nil, sportweekround, nil
	}
	defer rows.Close()
	for rows.Next() {
		var objSprotRound isg.SportRound
		err := rows.Scan(
			&objSprotRound.RoundID,
			&objSprotRound.Name,
			&objSprotRound.ShortRoundName,
			&objSprotRound.URL,
		)
		if err != nil {
			return nil, sportweekround, nil
		}
		objSprotRounds = append(objSprotRounds, objSprotRound)
	}

	return objSprotRounds, sportweekround, nil
}

// GetSportsRoundWeek :
func GetSportsRoundWeek(objsport isg.Sport, objleague isg.League, roundfilter string) ([]isg.SportRound, error) {

	var optValues []string
	var Objvalues string

	//var roundArr string
	if strings.HasPrefix(strings.ToLower(roundfilter), "round") {
		optValues = strings.SplitN(strings.ToLower(roundfilter), "-", 2)
		Objvalues = optValues[1]
	} else if strings.HasPrefix(strings.ToLower(roundfilter), "week") {
		optValues = strings.SplitN(strings.ToLower(roundfilter), "-", 2)
		Objvalues = optValues[1]
	} else if strings.HasPrefix(strings.ToLower(roundfilter), "seasontype") {
		optValues = strings.SplitN(strings.ToLower(roundfilter), "-", 2)
		Objvalues = optValues[1]
	} else if strings.HasPrefix(strings.ToLower(roundfilter), "matchday") {
		optValues = strings.SplitN(strings.ToLower(roundfilter), "-", 2)
		Objvalues = optValues[1]
	} else if strings.HasPrefix(strings.ToLower(roundfilter), "date") {
		return nil, nil
	} else if roundfilter != "" {
		Objvalues = strings.ToLower(roundfilter)
	}

	roundstr := strings.Replace(MakingRoundWeek(Objvalues), "+", "','", -1)
	roundstr = roundstr + "','" + strings.Replace(MakingRoundWeek(roundfilter), "+", "','", -1)

	objSprotRounds, sportweekround, err := GetRoundWeekDetails(objsport, objleague, roundstr)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(optValues) > 0 {
		if optValues[0] != sportweekround {
			return nil, errors.New("Invaild parameter: " + roundfilter)
		}
	}

	// for checking all ,playoff,reguler matches
	if objsport.SportID != "te" {
		roundweek := regexp.MustCompile(`all|pst|All|Reg|Pst|all`)
		if (len(Objvalues) == 4 && strings.Contains(Objvalues, "group")) || (len(Objvalues) == 8 && strings.Contains(Objvalues, "knockout")) ||
			(len(Objvalues) == 3 && strings.Contains(Objvalues, "reg")) {
			return nil, nil
		} else if roundweek.MatchString(Objvalues) {
			return nil, nil
		}
	}

	/*if len(strings.Split(roundfilter, "+")) != len(records) {
		return errors.New(sportweekround + " not found")
	}*/

	return objSprotRounds, nil
}

// MakingRoundWeek :
func MakingRoundWeek(optValues string) string {
	var valstr string

	arrayWeekRound := strings.Split(optValues, "+")
	for _, wekkRound := range arrayWeekRound {
		var weekroundval string

		val, err := strconv.Atoi(wekkRound)
		if err != nil {
			weekroundval = wekkRound
		} else {
			weekroundval = strconv.Itoa(val)
		}

		if valstr != "" {
			valstr = valstr + "+" + weekroundval
		} else {
			valstr = weekroundval
		}
	}
	return valstr
}

// GetSportsSeasonDetails :
func GetSportsSeasonDetails(objsport isg.Sport) []isg.Season {
	seasonIDs := []isg.Season{}
	var sqlstr string
	status := 1
	switch objsport.SportID {

	case "ar":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE  status = ? ORDER BY season_id DESC"

	case "af":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE   status = ? ORDER BY season_id DESC"

	case "bb":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + "  WHERE status = ? ORDER BY season_id DESC"

	case "sc":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ?  ORDER BY season_id DESC"

	case "te":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	case "rl":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	case "bl":
		sqlstr = "SELECT season_id, season,season FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	case "ih":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	case "ru":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	case "cr":
		sqlstr = "SELECT season_id, season,season_url FROM " + objsport.TableNameSeasons + " WHERE status = ? ORDER BY season_id DESC"

	}

	rows, err := SportsDb.Query(sqlstr, status)
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		record := isg.Season{}
		err := rows.Scan(&record.SeasonInternalID, &record.SeasonName, &record.SeasonURL)
		if err != nil {
			return nil
		}
		seasonIDs = append(seasonIDs, record)
	}

	return seasonIDs
}

//GetMatchCount :
func GetMatchCount(objsport isg.Sport, leagueID, seasonid int) (int, sql.NullInt64, error) {
	var sqlstr string
	var err error
	var totalPlayedMatch int
	var roundWeek sql.NullInt64
	status := "N"

	switch objsport.SportID {

	case "ar":
		sqlstr = "SELECT COUNT(*) AS playedmatch,MAX(round_id) AS round FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ?  "

	case "af":
		sqlstr = "SELECT COUNT(*) AS playedmatch, MAX(week_id) AS week FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	case "bb":
		if leagueID == 1 {
			sqlstr = "SELECT  COUNT(*) AS playedmatch, MAX(season_type_id) AS seasontype FROM " + objsport.TableNameMatches + "  WHERE  league_id = ? AND status = ? AND season_id = ? "
		} else {
			objsport.TableNameMatches = strings.Replace(objsport.TableNameMatches, "_daily_", "_round_", -1)
			sqlstr = "SELECT  COUNT(*) AS playedmatch, MAX(round_id) AS round FROM " + objsport.TableNameMatches + "  WHERE  league_id = ? AND status = ? AND season_id = ? "
		}
	case "sc":
		sqlstr = "SELECT  COUNT(*) AS playedmatch, MAX(week_id) AS week FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	case "te":
		sqlstr = "SELECT  COUNT(*) AS playedmatch FROM " + objsport.TableNameMatches + "  WHERE level_id = ? AND status = ? AND season_id = ? "

	case "rl":
		sqlstr = "SELECT  COUNT(*) AS playedmatch, MAX(round_id) AS round FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	case "bl":
		sqlstr = "SELECT  COUNT(*) AS playedmatch, MAX(round_id) AS round FROM " + objsport.TableNameMatches + "  WHERE league_id = ? AND status = ? AND season_id = ? "

	case "ih":
		sqlstr = "SELECT COUNT(*) AS playedmatch,  MAX(week_id) AS week FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	case "ru":
		sqlstr = "SELECT  COUNT(*) AS playedmatch,  MAX(round_id) AS round FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	case "cr":
		sqlstr = "SELECT  COUNT(*) AS playedmatch,  MAX(round_id) AS round FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ? AND season_id = ? "

	}
	//	fmt.Println(sqlstr)
	if objsport.SportID == "te" {
		err = SportsDb.QueryRow(sqlstr, leagueID, status, seasonid).Scan(&totalPlayedMatch)
	} else {
		err = SportsDb.QueryRow(sqlstr, leagueID, status, seasonid).Scan(&totalPlayedMatch, &roundWeek)
	}
	return totalPlayedMatch, roundWeek, err
}

//GetEventTeamID :
func GetEventTeamID(objsport isg.Sport, leagueID, matchID int) (isg.MatchInfo, error) {
	var sqlstr string
	var matchinfo isg.MatchInfo
	var matchtable string
	var homeTeamSBURL, awayTeamSBURL, matchDate, matchTime, counterDate, counterTime sql.NullString
	if objsport.SportID == "bb" && leagueID == 2 {
		matchtable = strings.Replace(objsport.TableNameMatches, "_daily_", "_round_", -1)
	} else {
		matchtable = objsport.TableNameMatches
	}
	switch objsport.SportID {
	case "ar":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_aussie_rules_round rounds ON rounds.round_id = matches.round_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	case "af":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,weeks.week_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_nfl_week weeks ON weeks.week_id = matches.week_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	case "sc":
		if leagueID == 10 || leagueID == 11 {
			sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,matchdays.match_day_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
				" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
				" LEFT JOIN isg_soccer_match_days matchdays ON matchdays.match_day_id = matches.week_id " +
				" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
				" WHERE league_id = ? AND  match_id = ? "
		} else if leagueID == 17 {
			sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,weeks.week_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
				" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
				" LEFT JOIN isg_soccer_worldcup_week weeks ON weeks.week_id = matches.week_id " +
				" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
				" WHERE league_id = ? AND  match_id = ? "
		} else {
			sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,weeks.week_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
				" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
				" LEFT JOIN isg_soccer_week weeks ON weeks.week_id = matches.week_id " +
				" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
				" WHERE league_id = ? AND  match_id = ? "
		}
	case "rl":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_rugby_league_round rounds ON rounds.round_id = matches.round_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	case "ru":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_rugby_union_round rounds ON rounds.round_id = matches.round_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	case "ih":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,weeks.week_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_hockey_week weeks ON weeks.week_id = matches.week_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE league_id = ? AND  match_id = ? "
	case "bl":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_baseball_round rounds ON rounds.round_id = matches.round_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	case "bb":
		if leagueID == 1 {
			sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,seasontype.type_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
				" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
				" LEFT JOIN isg_basketball_season_type seasontype ON seasontype.season_type_id = matches.season_type_id " +
				" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
				" WHERE matches.league_id = ? AND  matches.match_id = ? "
		} else {
			sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
				" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
				" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
				" LEFT JOIN isg_basketball_round rounds ON rounds.round_id = matches.round_id " +
				" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
				" WHERE matches.league_id = ? AND  matches.match_id = ? "
		}
	case "te":
		sqlstr = "SELECT matches.match_id,matches.player1_id,player1.isg_api_id,IFNULL(player1.filter_name,''),IFNULL(player1.full_name,''),IFNULL(player1.short_name,''),IFNULL(player1.player_url_name,''),IFNULL(country1.flag,''),matches.player2_id,player2.isg_api_id,IFNULL(player2.filter_name,''),IFNULL(player2.full_name,''),IFNULL(player2.short_name,''),IFNULL(player2.player_url_name,''),IFNULL(country2.flag,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,tournament.tournament_url,rounds.round_url,tournament.filter_name,country.country,matches.match_date,matches.counter_date,matches.counter_time,matches.status  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_tennis_players player1 ON player1.player_id = matches.player1_id " +
			" LEFT JOIN isg_tennis_players player2 ON player2.player_id = matches.player2_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" LEFT JOIN isg_country country1 ON country1.country_id = player1.country_id " +
			" LEFT JOIN isg_country country2 ON country2.country_id = player2.country_id " +
			" LEFT JOIN isg_tennis_tournament tournament ON tournament.tournament_id = matches.tournament_id " +
			" LEFT JOIN isg_tennis_tournament_season_round AS seasonround ON seasonround.match_id = matches.match_id" +
			" LEFT JOIN isg_tennis_round AS rounds ON rounds.round_id = seasonround.round_id" +
			" LEFT JOIN isg_tennis_tournament_season_venue season_venue ON season_venue.tournament_id = matches.tournament_id AND season_venue.level_id = matches.level_id AND season_venue.season_id = matches.season_id" +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = season_venue.venue_id" +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country" +
			" WHERE matches.level_id = ? AND  matches.match_id = ? "

	case "cr":
		sqlstr = "SELECT matches.match_id,matches.home_team_id,hometeam.isg_api_id,IFNULL(hometeam.filtername,''),IFNULL(hometeam.team_name,''),IFNULL(hometeam.abbreviation,''),IFNULL(hometeam.url,''),IFNULL(hometeam.icon,''),matches.away_team_id,hometeam.isg_api_id,IFNULL(awayteam.filtername,''),IFNULL(awayteam.team_name,''),IFNULL(awayteam.abbreviation,''),IFNULL(awayteam.url,''),IFNULL(awayteam.icon,''),matches.season_id,IFNULL(seasons_tb.season,''),matches.match_time,rounds.round_url,matches.match_date,hometeam.sportsbet_url,awayteam.sportsbet_url,matches.counter_date,matches.counter_time,hometeam.short_teamname,awayteam.short_teamname,matches.status,hometeam.isg_api_name,awayteam.isg_api_name  FROM " + matchtable + " AS matches  " +
			" LEFT JOIN isg_team hometeam ON hometeam.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team awayteam ON awayteam.team_id = matches.away_team_id " +
			" LEFT JOIN isg_cricket_round rounds ON rounds.round_id = matches.round_id " +
			" LEFT JOIN " + objsport.TableNameSeasons + " seasons_tb ON seasons_tb.season_id = matches.season_id " +
			" WHERE matches.league_id = ? AND  matches.match_id = ? "
	}
	if objsport.SportID == "te" {
		err := SportsDb.QueryRow(sqlstr, leagueID, matchID).Scan(&matchinfo.MatchID, &matchinfo.HomeTeamInternalID, &matchinfo.HomeTeamID, &matchinfo.HomeTeamFilterName, &matchinfo.HomeTeamName, &matchinfo.HomeTeamAbbr, &matchinfo.HomeTeamURL, &matchinfo.HomeTeamIcon, &matchinfo.AwayTeamInternalID, &matchinfo.AwayTeamID, &matchinfo.AwayTeamFilterName, &matchinfo.AwayTeamName, &matchinfo.AwayTeamAbbr, &matchinfo.AwayTeamURL, &matchinfo.AwayTeamIcon, &matchinfo.SeasonID, &matchinfo.Season, &matchTime, &matchinfo.TournamentURL, &matchinfo.RoundURL, &matchinfo.TournamentFilterName, &matchinfo.TournamentCountryName, &matchDate, &counterDate, &counterTime, &matchinfo.Status)
		if err != nil {
			return matchinfo, err
		}
	} else {
		err := SportsDb.QueryRow(sqlstr, leagueID, matchID).Scan(&matchinfo.MatchID, &matchinfo.HomeTeamInternalID, &matchinfo.HomeTeamID, &matchinfo.HomeTeamFilterName, &matchinfo.HomeTeamName, &matchinfo.HomeTeamAbbr, &matchinfo.HomeTeamURL, &matchinfo.HomeTeamIcon, &matchinfo.AwayTeamInternalID, &matchinfo.AwayTeamID, &matchinfo.AwayTeamFilterName, &matchinfo.AwayTeamName, &matchinfo.AwayTeamAbbr, &matchinfo.AwayTeamURL, &matchinfo.AwayTeamIcon, &matchinfo.SeasonID, &matchinfo.Season, &matchTime, &matchinfo.RoundURL, &matchDate, &homeTeamSBURL, &awayTeamSBURL, &counterDate, &counterTime, &matchinfo.HomeTeamShortName, &matchinfo.AwayTeamShortName, &matchinfo.Status, &matchinfo.HomeTeamNickName, &matchinfo.AwayTeamNickName)
		if err != nil {
			return matchinfo, err
		}
		if homeTeamSBURL.Valid && awayTeamSBURL.Valid {
			matchinfo.HomeTeamSBURL = homeTeamSBURL.String
			matchinfo.AwayTeamSBURL = awayTeamSBURL.String
		}
	}
	matchinfo.MatchDate = matchDate.String
	matchinfo.MatchTime = matchTime.String
	matchinfo.LocalDate = counterDate.String
	matchinfo.LocalTime = counterTime.String

	return matchinfo, nil
}

//GetTeamTotalMatch :
func GetTeamTotalMatch(objsport isg.Sport, leagueID, seasonid, team1id, team2id int) (int, error) {
	var sqlstr string
	var matchtable string
	var totalcnt sql.NullInt64
	status := "N"
	if objsport.SportID == "bb" && leagueID == 2 {
		matchtable = strings.Replace(objsport.TableNameMatches, "_daily_", "_round_", -1)
	} else {
		matchtable = objsport.TableNameMatches
	}
	switch objsport.SportID {
	case "ar", "af", "sc", "rl", "ru", "ih", "bl", "bb", "cr":
		sqlstr = "SELECT  COUNT(*) AS playedteammatch FROM " + matchtable +
			" WHERE league_id = ? AND status = ? AND season_id = ? AND ((home_team_id = ? OR home_team_id = ?) OR (away_team_id = ? OR away_team_id = ?))"
	case "te":
		sqlstr = "SELECT  COUNT(*) AS playedteammatch FROM " + matchtable +
			" WHERE level_id = ? AND status = ? AND season_id = ? AND ((player1_id = ? OR player1_id = ?) OR (player2_id = ? OR player2_id = ?))"
	}
	err := SportsDb.QueryRow(sqlstr, leagueID, status, seasonid, team1id, team2id, team1id, team2id).Scan(&totalcnt)
	totalplayed := int(totalcnt.Int64)
	if err != nil {
		return totalplayed, err
	}
	return totalplayed, nil
}

// GetEventLeagueDetails :
func GetEventLeagueDetails(sportid int, leagueid int) (isg.League, error) {
	league := isg.League{}
	var sqlstr string
	switch sportid {

	case 1: // Aussie Rules
		sqlstr = "SELECT league_id,league_name,league_url FROM isg_aussie_rules_leagues WHERE  league_id = ? "
	case 2: //American Football
		sqlstr = "SELECT league_id, first_name,league_url FROM isg_american_football_league WHERE   league_id = ?  "

	case 3: //Basketball
		sqlstr = "SELECT league_id, league_name,league_url FROM isg_basketball_leagues WHERE  league_id = ?  "

	case 4: // Soccer
		sqlstr = "SELECT league_id, first_name,league_url FROM isg_soccer_leagues WHERE  league_id = ? "

	case 5: // Cricket
		sqlstr = "SELECT league_id, league_name, league_url FROM isg_cricket_leagues WHERE  league_id = ? "

	case 6: // Tennis
		sqlstr = "SELECT level_id,level_name,level_url FROM isg_tennis_tournament_level WHERE  level_id = ?"

	case 7: // Rugby League
		sqlstr = "SELECT league_id, league_name,league_url FROM isg_rugby_league WHERE  league_id = ? "

	case 8: // Hockey
		sqlstr = "SELECT league_id,league_name,league_url FROM isg_hockey_leagues WHERE  league_id = ? "

	case 9: // Baseball
		sqlstr = "SELECT league_id,league_name,league_url FROM isg_baseball_leagues WHERE  league_id = ?"

	case 10: // Rugby Union
		sqlstr = "SELECT league_id, league_name,league_url FROM isg_rugby_union_league WHERE  league_id = ? "
	}

	err := SportsDb.QueryRow(sqlstr, leagueid).Scan(&league.LeagueInternalID, &league.LeagueName, &league.LeagueURL)

	return league, err

}

//GetTeamDetails :
func GetTeamDetails(teamid string) (isg.Team, error) {
	objteam := isg.Team{}
	query := "SELECT team_id,team_name, abbreviation,team_color,flag,isg_api_id,icon,filtername,url FROM isg_team WHERE (team_id=? OR team_name=? OR filtername=? OR abbreviation=? OR isg_api_id=? OR url=?)  AND status = ?"
	err := SportsDb.QueryRow(query, teamid, teamid, teamid, teamid, teamid, teamid, `1`).Scan(&objteam.TeamInternalID, &objteam.FullName, &objteam.Abbreviation, &objteam.TeamColor, &objteam.TeamFlag, &objteam.TeamID, &objteam.TeamFlag, &objteam.TeamName, &objteam.TeamURL)
	if err != nil {
		return objteam, err
	}
	return objteam, nil
}

//MakeForm :
func MakeForm(forms []string) string {
	var formlen int
	var form []string
	if len(forms) >= 5 {
		formlen = 5
	} else {
		formlen = len(forms)
	}
	for i := 0; i < formlen; i++ {
		form = append(form, forms[i])
	}
	return strings.Join(Reverse(form), "")
}

//Reverse :
func Reverse(strform []string) []string {
	for i := 0; i < len(strform)/2; i++ {
		j := len(strform) - i - 1
		strform[i], strform[j] = strform[j], strform[i]
	}
	return strform
}

//GetLeagueTeamDetails :
func GetLeagueTeamDetails(teamid string, sportobj isg.Sport, leagueID int) (isg.Team, error) {
	objteam := isg.Team{}
	var err error
	switch sportobj.SportID {
	case "te":
		query := "SELECT player_id,full_name, isg_api_id,filter_name, player_url_name, flag FROM isg_tennis_players " +
			"LEFT JOIN isg_country ON isg_tennis_players.country_id = isg_country.country_id  " +
			" WHERE (player_url_name=? OR filter_name=? OR isg_api_id=? )  AND isg_tennis_players.status = ? AND level_id = ?"
		err = SportsDb.QueryRow(query, teamid, teamid, teamid, `1`, leagueID).Scan(
			&objteam.TeamInternalID,
			&objteam.FullName,
			&objteam.TeamID,
			&objteam.TeamName,
			&objteam.TeamURL,
			&objteam.TeamFlag,
		)

	default:
		if sportobj.SportInternalID == 3 {
			if leagueID != 1 {
				sportobj.TableNameMatches = "isg_basketball_round_matches"
			}
		}
		sqlstr := "SELECT team.team_id,team_name, abbreviation, team_color, flag, team.isg_api_id, icon, filtername, url " +
			" FROM isg_team  as team " +
			" LEFT JOIN " + sportobj.TableNameMatches + " as matches ON (matches.home_team_id = team.team_id OR matches.away_team_id = team.team_id) " +
			" WHERE (team_name = ? OR filtername = ? OR short_teamname = ? OR team.isg_api_regionname = ? OR team.isg_api_name = ? OR team.isg_api_id = ? OR team.url = ?)  " +
			"AND team.status = ? AND league_id = ? AND sport_id  = ? " +
			" Group by team.team_id "

		err = SportsDb.QueryRow(sqlstr, teamid, teamid, teamid, teamid, teamid, teamid, teamid, `1`, leagueID, sportobj.SportInternalID).Scan(&objteam.TeamInternalID, &objteam.FullName, &objteam.Abbreviation, &objteam.TeamColor, &objteam.TeamFlag, &objteam.TeamID, &objteam.Pitcher, &objteam.TeamName, &objteam.TeamURL)
	}
	//
	if err != nil {
		return objteam, err
	}
	return objteam, nil
}

//ValidateConference :
func ValidateConference(conference string, sportobj isg.Sport) (int, error) {
	var count int
	query := "SELECT COUNT(1) FROM isg_team WHERE sport_id= ? AND conference = ?"
	err := SportsDb.QueryRow(query, sportobj.SportInternalID, conference).Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

//ValidateDivision :
func ValidateDivision(division string, sportobj isg.Sport) (int, error) {
	var count int
	query := "SELECT COUNT(1) FROM isg_team WHERE sport_id= ? AND division = ?"
	err := SportsDb.QueryRow(query, sportobj.SportInternalID, division).Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

//GetSportsFormStreakData :
func GetSportsFormStreakData(sportid string, leagueid int, teamid, season, tablename string) ([]isg.SportsFormStreak, error) {
	var formDatas []isg.SportsFormStreak
	seasonArr := strings.Split(season, ",")
	var sqlstr string

	switch sportid {
	case "te":
		sqlstr = "SELECT form, streak, total_games  " +
			" FROM " + tablename + " WHERE player_id = ? AND level_id = ? AND season_id <= ? GROUP BY season_id Order By season_id ASC "

	case "ih":
		sqlstr = "SELECT form, streak, matches  " +
			" FROM " + tablename + " WHERE team_id = ?  AND league_id = ? AND season_id <= ? Order By season_id ASC "

	default:
		sqlstr = "SELECT form, streak, total_match  " +
			" FROM " + tablename + " WHERE team_id = ? AND league_id = ? AND season_id <= ? Order By season_id ASC "

	}

	rows, err := SportsDb.Query(sqlstr, teamid, leagueid, seasonArr[len(seasonArr)-1])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		formData := isg.SportsFormStreak{}
		err := rows.Scan(
			&formData.Form,
			&formData.Streak,
			&formData.MatchCount,
		)
		if err != nil {
			return nil, err
		}
		formDatas = append(formDatas, formData)
	}
	return formDatas, nil
}

//CalculationBTKSFormStreak :
func CalculationBTKSFormStreak(objSport isg.Sport, records []isg.SportsFormStreak) (string, string) {

	var count int
	var finalStreak string
	var finalForm string
	totalLength := len(records) - 1

	if len(records) > 1 {
		for i := totalLength; i > 0; i-- {

			streak1 := (records[i].Streak)[0]
			streak2 := (records[i-1].Streak)[0]
			if objSport.SportInternalID == 8 || objSport.SportInternalID == 5 {
				if streak1 == 'l' || streak1 == 'L' {
					streak1 = 'L'
				}
				if streak2 == 'W' || streak2 == 'w' {
					streak2 = 'W'
				}
				if streak1 == 'W' || streak1 == 'w' {
					streak1 = 'W'
				}
				if streak2 == 'l' || streak2 == 'L' {
					streak2 = 'L'
				}
			}

			streakcount, _ := strconv.Atoi(string(records[i].Streak)[1:])
			if len(records[i].Form) == streakcount && streak1 == streak2 {

				if i == totalLength && streakcount != records[i].MatchCount {
					finalStreak = records[i].Streak
					break
				}

				if i == totalLength {
					x, _ := strconv.Atoi(string((records[i].Streak)[1:]))
					y, _ := strconv.Atoi(string((records[i-1].Streak)[1:]))
					count = x + y
				} else {
					y, _ := strconv.Atoi(string((records[i-1].Streak)[1]))
					count = count + y
					finalStreak = string(streak1) + "" + strconv.Itoa(count)
				}

			} else {
				if i == totalLength {
					finalStreak = records[i].Streak
				} else {
					finalStreak = string(streak1) + "" + strconv.Itoa(count)
				}
				break
			}
		}
	} else {
		if len(records) != 0 {
			if objSport.SportInternalID == 8 || objSport.SportInternalID == 5 {
				if objSport.SportInternalID == 8 && (records[0].Streak)[0] == 'l' {
					records[0].Streak = "OTL" + string(records[0].Streak[1])
				} else if objSport.SportInternalID == 5 && (records[0].Streak)[0] == 'l' {
					records[0].Streak = "l" + string(records[0].Streak[1])
				}
			}
			return records[0].Form, records[0].Streak
		}
		return finalForm, finalStreak
	}
	if objSport.SportInternalID == 8 || objSport.SportInternalID == 5 {
		if len(records) != 0 {
			if objSport.SportInternalID == 8 && (records[totalLength].Streak)[0] == 'l' {

				finalStreak = "OTL" + string(finalStreak[1])

			} else if objSport.SportInternalID == 5 && (records[totalLength].Streak)[0] == 'l' {

				finalStreak = "l" + string(finalStreak[1])
			}
		}
	}
	//calculate form
	for _, formdata := range records {
		finalForm = finalForm + formdata.Form
	}
	if len(finalForm) > 5 {
		finalForm = finalForm[len(finalForm)-5:]
	}
	return finalForm, finalStreak
}

//GetCurrentSeasonid : get current season ID according to league
func GetCurrentSeasonid(objsport isg.Sport, leagueID int) (int, error) {
	var seasonID int
	var err error
	status := "N"
	switch objsport.SportID {
	case "te":
		sqlstr := "SELECT season_id FROM " + objsport.TableNameMatches + " WHERE level_id = ? AND status = ? ORDER BY season_id DESC LIMIT 0,1"
		err = SportsDb.QueryRow(sqlstr, leagueID, status).Scan(&seasonID)
	default:
		if objsport.SportInternalID == 3 && leagueID == 2 {
			objsport.TableNameMatches = "isg_basketball_round_matches"
		}
		sqlstr := "SELECT season_id FROM " + objsport.TableNameMatches + " WHERE league_id = ? AND status = ?  ORDER BY season_id DESC LIMIT 0,1"
		err = SportsDb.QueryRow(sqlstr, leagueID, status).Scan(&seasonID)
	}
	if err != nil {
		return seasonID, err
	}
	return seasonID, nil
}

// GetSportsCurrentRound :
func GetSportsCurrentRound(objsport isg.Sport, objleague isg.League, seasonID int) (isg.RoundWeek, error) {

	var roundweek isg.RoundWeek
	var datediff, timediff *int
	var sqlstr string
	switch objsport.SportInternalID {

	case 1: // Aussie Rules
		sqlstr = "SELECT b.round_id, short_round_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_aussie_rules_matches AS matches " +
			" LEFT JOIN isg_aussie_rules_round AS b ON b.round_id = matches.round_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "

	case 2: //American Football
		sqlstr = "SELECT b.week_id, short_week_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_nflmatches AS matches" +
			" LEFT JOIN isg_nfl_week AS b ON b.week_id = matches.week_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "

	case 3: //Basketball
		sqlstr = "SELECT b.season_type_id, short_type_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_basketball_daily_matches AS matches " +
			" LEFT JOIN isg_basketball_season_type AS b ON b.season_type_id = matches.season_type_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "
		if objleague.LeagueInternalID == 2 {
			sqlstr = "SELECT b.round_id, short_round_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
				" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
				" FROM isg_basketball_round_matches AS matches " +
				" LEFT JOIN isg_basketball_round AS b ON b.round_id = matches.round_id " +
				" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "
		}

	case 4: // Soccer
		sqlstr = "SELECT b.week_id, short_week_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_soccermatches AS matches" +
			" LEFT JOIN isg_soccer_week AS b ON b.week_id = matches.week_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "

	case 6: // Tennis

	case 7: // Rugby League
		sqlstr = "SELECT b.round_id, short_round_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_rugby_league_matches AS matches" +
			" LEFT JOIN isg_rugby_league_round AS b ON b.round_id = matches.round_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "

	case 8: // Hockey
		sqlstr = "SELECT b.week_id, short_week_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_hockeymatches AS matches " +
			" LEFT JOIN isg_hockey_week AS b ON b.week_id = matches.week_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "

	case 9: // Baseball
		sqlstr = "SELECT b.round_id, round_short_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_baseball_matches AS matches " +
			" LEFT JOIN isg_baseball_round AS b ON b.round_id = matches.round_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "
	case 10: // Rugby Union
		sqlstr = "SELECT b.round_id, short_round_name, ABS(DATEDIFF(counter_date, '" + time.Now().In(AEST).Format("2006-01-02") + "')) as date_diff, " +
			" ABS(time(counter_time) - time('" + time.Now().In(AEST).Format("15:04:05") + "')) as time_diff " +
			" FROM isg_rugby_union_matches AS matches " +
			" LEFT JOIN isg_rugby_union_round AS b ON b.round_id = matches.round_id " +
			" WHERE league_id = ? AND season_id = ? AND counter_date is not null AND counter_time is not null ORDER BY matches.status DESC, date_diff ASC, time_diff ASC LIMIT 0,1 "
	}
	err := SportsDb.QueryRow(sqlstr, objleague.LeagueInternalID, seasonID).Scan(&roundweek.RoundWeekID, &roundweek.RoundWeekName, &datediff, &timediff)
	if err != nil {
		return roundweek, err
	}

	return roundweek, nil

}

//GetTennisAllTournamentsDetails : according to level id
func GetTennisAllTournamentsDetails(levelID int) ([]isg.Tournament, error) {
	var objtournament []isg.Tournament
	var country, city *string
	sqlstr := `SELECT distinct(tour.tournament_id),tour.isg_api_id,tour.filter_name,tour.tournament_name,country.country,venue.city
		FROM isg_tennis_tournament as tour
		LEFT JOIN isg_tennis_tournament_season_venue as tvenue ON tvenue.tournament_id = tour.tournament_id
		LEFT JOIN isg_venue as venue ON venue.venue_id = tvenue.venue_id
		LEFT JOIN isg_country as country ON country.country_id = venue.country
		WHERE tour.level_id = ?`
	rows, err := SportsDb.Query(sqlstr, levelID)
	if err != nil {
		return objtournament, err
	}
	defer rows.Close()
	for rows.Next() {
		objtd := isg.Tournament{}
		err := rows.Scan(
			&objtd.TournamentInternalID,
			&objtd.TournamentID,
			&objtd.TournamentFilterName,
			&objtd.TournamentName,
			&country,
			&city,
		)

		if country != nil {
			objtd.TournamentCountry = *country
		}
		if city != nil {
			objtd.TournamentCity = *city
		}

		if err != nil {
			return objtournament, err
		}
		objtournament = append(objtournament, objtd)
	}
	return objtournament, nil
}

//CheckCacheKeyString :
func CheckCacheKeyString(str string) string {
	if str != "" {
		return ":" + str
	}
	return ""
}

// GetSportsMatchTips :
func GetSportsMatchTips(tableName string, sportID, matchID, leagueID, providerID int) ([]isg.Tips, error) {
	var objtipsInfo []isg.Tips

	league := " AND league_id = ? "
	if sportID == 6 {
		league = " AND level_id = ? "
	}

	sqlstr := "SELECT tips.title, COALESCE(tip, ''), COALESCE(roughie, ''), options.option_1, options.option_2 " +
		" FROM " + tableName + " AS tips " +
		" LEFT JOIN isg_tips_options options ON options.sport_id = ? AND options.provider_id = ?  " +
		" WHERE match_id = ? " + league + "  AND tips.provider_id = ? "

	rows, err := SportsDb.Query(sqlstr, sportID, providerID, matchID, leagueID, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {

		var objtips isg.Tips
		err := rows.Scan(
			&objtips.TipsTitle,
			&objtips.Tips,
			&objtips.Roughie,
			&objtips.Option1,
			&objtips.Option2,
		)
		if err != nil {
			return nil, err
		}
		objtipsInfo = append(objtipsInfo, objtips)
	}
	return objtipsInfo, nil
}

//GetLeagueCountryDetails :
func GetLeagueCountryDetails(country string, sportobj isg.Sport) (int, error) {
	var countryID int

	sqlstr := "SELECT coun.country_id " +
		" FROM isg_venue venue " +
		" LEFT JOIN isg_country coun ON coun.country_id = venue.country" +
		" WHERE venue.sports = ? AND coun.country_url = '" + country + "'" +
		" LIMIT 1"

	err := SportsDb.QueryRow(sqlstr, sportobj.SportInternalID).Scan(&countryID)
	if err != nil {
		return countryID, err
	}
	return countryID, nil
}

//GetVenueDetails :
func GetVenueDetails(venue string, sportobj isg.Sport) (isg.Venue, error) {
	objVenue := isg.Venue{}

	query := "SELECT venue_id FROM isg_venue WHERE (venue=? OR filtername=? OR friendlyname=? OR isg_api_id=? OR venue_id=?) AND sports = ? "
	err := SportsDb.QueryRow(query, venue, venue, venue, venue, venue, sportobj.SportInternalID).Scan(&objVenue.VenueID)
	//
	if err != nil {
		return objVenue, err
	}
	return objVenue, nil
}

//GetRegionDetails
func GetRegionDetails(region string) (string, error) {
	var regionId string
	query := "SELECT region_id FROM isg_regions WHERE (region_id=? OR region_name=?) "
	err := SportsDb.QueryRow(query, region, region).Scan(&regionId)
	if err != nil {
		return "0", err
	}

	return regionId, nil
}

//MakingLiveOddsSort :
func MakingLiveOddsSort(liveOdds []isg.FixtureOdds, sportID int) []isg.FixtureOdds {
	var fixtureLiveOdds []isg.FixtureOdds
	var fixtureLiveOdd isg.FixtureOdds
	var homeliveOdd, awayliveOdd, drawliveOdd *float64
	var flagClosing, flagHomeline, flagBtts bool

	for i, odds := range liveOdds {

		if odds.HomeOdds != nil && odds.AwayOdds != nil {
			//home odds
			if homeliveOdd == nil || *odds.HomeOdds > *homeliveOdd {
				homeliveOdd = odds.HomeOdds
			}

			//away odds
			if awayliveOdd == nil || *odds.AwayOdds > *awayliveOdd {
				awayliveOdd = odds.AwayOdds
			}

			if sportID == 4 {
				//draw odds
				if drawliveOdd == nil || *odds.DrawOdds > *drawliveOdd {
					drawliveOdd = odds.DrawOdds
				}
			}
			fixtureLiveOdd.MarketName = "Win"

		}

		//closing
		if odds.ClosingTotal != nil && !flagClosing && sportID != 5 {
			flagClosing = true
			fixtureLiveOdd.ClosingTotal = odds.ClosingTotal
			fixtureLiveOdd.OverOdds = odds.OverOdds
			fixtureLiveOdd.UnderOdds = odds.UnderOdds
			fixtureLiveOdd.MarketName = "Total"
		}
		//homeline
		if odds.HomeLine != nil && !flagHomeline && sportID != 5 {
			flagHomeline = true
			fixtureLiveOdd.HomeLine = odds.HomeLine
			fixtureLiveOdd.HomeLineOdds = odds.HomeLineOdds
			fixtureLiveOdd.AwayLine = odds.AwayLine
			fixtureLiveOdd.AwayLineOdds = odds.AwayLineOdds
			fixtureLiveOdd.MarketName = "Line"
		}
		if sportID == 4 {
			//btts odds
			if odds.BttsYesOdds != nil && !flagBtts {
				flagBtts = true
				fixtureLiveOdd.BttsYesOdds = odds.BttsYesOdds
				fixtureLiveOdd.BttsNoOdds = odds.BttsNoOdds
			}
		}

		if (len(liveOdds)-1) == i || odds.MatchID != liveOdds[i+1].MatchID {
			fixtureLiveOdd.MatchID = odds.MatchID
			fixtureLiveOdd.HomeOdds = homeliveOdd
			fixtureLiveOdd.AwayOdds = awayliveOdd
			if sportID == 4 {
				fixtureLiveOdd.DrawOdds = drawliveOdd
			}
			fixtureLiveOdd.ProviderInfo.Name = odds.ProviderInfo.Name
			fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveOdd)
			fixtureLiveOdd = isg.FixtureOdds{}
			homeliveOdd, awayliveOdd, drawliveOdd = nil, nil, nil
			flagHomeline, flagBtts, flagClosing = false, false, false
		}
	}

	return fixtureLiveOdds
}

//MakingFixtureLiveOddsSort :
func MakingFixtureLiveOddsSort(liveOdds []isg.FixtureOdds, sportID int) []isg.GeniusOddsMarket {
	var fixtureLiveOdds []isg.GeniusOddsMarket
	var fixtureLiveOdd, fixtureLiveOddHome, fixtureLiveOddAway, fixtureLiveLineHome, fixtureLiveLineAway, fixtureLiveOver, fixtureLiveUnder, fixtureLiveBttsYes, fixtureLiveBttsNo isg.GeniusOddsMarket
	var homeliveOdd, awayliveOdd *float64
	var homeliveOddProvider, awayliveOddProvider, homeliveOddProviderIcon, awayliveOddProviderIcon string
	var flagHomeline, flagClosing, flagBtts bool

	for i, odds := range liveOdds {

		if odds.HomeOdds != nil && odds.AwayOdds != nil {
			//home odds
			if homeliveOdd == nil || *odds.HomeOdds > *homeliveOdd {
				homeliveOdd = odds.HomeOdds
				homeliveOddProvider = odds.ProviderInfo.Name
				homeliveOddProviderIcon = odds.ProviderInfo.Icon
			}

			//away odds
			if awayliveOdd == nil || *odds.AwayOdds > *awayliveOdd {
				awayliveOdd = odds.AwayOdds
				awayliveOddProvider = odds.ProviderInfo.Name
				awayliveOddProviderIcon = odds.ProviderInfo.Icon
			}

		}

		if sportID == 4 {
			if odds.BttsYesOdds != nil && !flagBtts {
				flagBtts = true
				fixtureLiveOdd.MatchID = odds.MatchID
				fixtureLiveBttsYes.MarketName = "BTTS"
				fixtureLiveBttsYes.GroupName = "BttsYes"
				fixtureLiveBttsYes.MatchID = odds.MatchID
				fixtureLiveBttsYes.MarketPrice = odds.BttsYesOdds
				fixtureLiveBttsYes.ProviderInfo.Name = odds.ProviderInfo.Name
				fixtureLiveBttsYes.ProviderInfo.Icon = odds.ProviderInfo.Icon

				fixtureLiveBttsNo.MarketName = "BTTS"
				fixtureLiveBttsNo.GroupName = "BttsNo"
				fixtureLiveBttsNo.MatchID = odds.MatchID
				fixtureLiveBttsNo.MarketPrice = odds.BttsNoOdds
				fixtureLiveBttsNo.ProviderInfo.Name = odds.ProviderInfo.Name
				fixtureLiveBttsNo.ProviderInfo.Icon = odds.ProviderInfo.Icon
			}
		}

		//closing
		if odds.ClosingTotal != nil && !flagClosing && sportID != 5 {
			flagClosing = true
			fixtureLiveOdd.MatchID = odds.MatchID
			fixtureLiveOver.MarketName = "Total"
			fixtureLiveOver.GroupName = "Over"
			fixtureLiveOver.MatchID = odds.MatchID
			fixtureLiveOver.MarketPrice = odds.OverOdds
			fixtureLiveOver.MarketFlucPrice = odds.ClosingTotal
			fixtureLiveOver.ProviderInfo.Name = odds.ProviderInfo.Name
			fixtureLiveOver.ProviderInfo.Icon = odds.ProviderInfo.Icon

			fixtureLiveUnder.MarketName = "Total"
			fixtureLiveUnder.GroupName = "Under"
			fixtureLiveUnder.MatchID = odds.MatchID
			fixtureLiveUnder.MarketPrice = odds.UnderOdds
			fixtureLiveUnder.MarketFlucPrice = odds.ClosingTotal
			fixtureLiveUnder.ProviderInfo.Name = odds.ProviderInfo.Name
			fixtureLiveUnder.ProviderInfo.Icon = odds.ProviderInfo.Icon
		}

		//homeline
		if odds.HomeLine != nil && !flagHomeline && sportID != 5 {
			flagHomeline = true
			fixtureLiveOdd.MatchID = odds.MatchID
			fixtureLiveLineHome.MarketName = "Line"
			fixtureLiveLineHome.GroupName = "Home"
			fixtureLiveLineHome.MatchID = odds.MatchID
			fixtureLiveLineHome.MarketPrice = odds.HomeLineOdds
			fixtureLiveLineHome.MarketFlucPrice = odds.HomeLine
			fixtureLiveLineHome.ProviderInfo.Name = odds.ProviderInfo.Name
			fixtureLiveLineHome.ProviderInfo.Icon = odds.ProviderInfo.Icon

			fixtureLiveLineAway.MarketName = "Line"
			fixtureLiveLineAway.GroupName = "Away"
			fixtureLiveLineAway.MatchID = odds.MatchID
			fixtureLiveLineAway.MarketPrice = odds.AwayLineOdds
			fixtureLiveLineAway.MarketFlucPrice = odds.AwayLine
			fixtureLiveLineAway.ProviderInfo.Name = odds.ProviderInfo.Name
			fixtureLiveLineAway.ProviderInfo.Icon = odds.ProviderInfo.Icon

		}

		if (len(liveOdds)-1) == i || odds.MatchID != liveOdds[i+1].MatchID {
			fixtureLiveOdd.MatchID = odds.MatchID

			if homeliveOdd != nil {
				fixtureLiveOddHome.MarketName = "Win"
				fixtureLiveOddHome.GroupName = "Home"
				fixtureLiveOddHome.MatchID = odds.MatchID
				fixtureLiveOddHome.MarketPrice = homeliveOdd
				fixtureLiveOddHome.ProviderInfo.Name = homeliveOddProvider
				fixtureLiveOddHome.ProviderInfo.Icon = homeliveOddProviderIcon
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveOddHome)

				fixtureLiveOddAway.MarketName = "Win"
				fixtureLiveOddAway.GroupName = "Away"
				fixtureLiveOddAway.MatchID = odds.MatchID
				fixtureLiveOddAway.MarketPrice = awayliveOdd
				fixtureLiveOddAway.ProviderInfo.Name = awayliveOddProvider
				fixtureLiveOddAway.ProviderInfo.Icon = awayliveOddProviderIcon
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveOddAway)
			}

			// for line
			if fixtureLiveLineHome.MarketPrice != nil {
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveLineHome)
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveLineAway)
			}

			// total
			if fixtureLiveOver.MarketPrice != nil {
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveOver)
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveUnder)
			}

			// btts
			if fixtureLiveBttsYes.MarketPrice != nil {
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveBttsYes)
				fixtureLiveOdds = append(fixtureLiveOdds, fixtureLiveBttsNo)
			}

			homeliveOdd, awayliveOdd = nil, nil
			homeliveOddProvider, awayliveOddProvider, homeliveOddProviderIcon, awayliveOddProviderIcon = "", "", "", ""
			flagHomeline, flagClosing, flagBtts = false, false, false
			fixtureLiveOddHome, fixtureLiveOddAway, fixtureLiveLineHome, fixtureLiveLineAway = isg.GeniusOddsMarket{}, isg.GeniusOddsMarket{}, isg.GeniusOddsMarket{}, isg.GeniusOddsMarket{}
			fixtureLiveOver, fixtureLiveUnder = isg.GeniusOddsMarket{}, isg.GeniusOddsMarket{}
			fixtureLiveBttsYes, fixtureLiveBttsNo = isg.GeniusOddsMarket{}, isg.GeniusOddsMarket{}
		}

	}

	return fixtureLiveOdds
}

// GenerateSQLQueryForGeniusOdds :
func GenerateSQLQueryForGeniusOdds(objSport isg.Sport, objLeague isg.League, matchID int) string {

	var _sqlstr, _searchStr, limitstr string

	if matchID != 0 {
		_searchStr = " matches.match_id = " + strconv.Itoa(matchID) + " AND "
	}

	if matchID == 0 {
		_searchStr = "concat(matches.counter_date, ' ', matches.counter_time) >= '" + time.Now().In(AEST).Format("2006-01-02 15:04:05") + "' AND "
		limitstr = " LIMIT 0, 10 "
	}
	switch objSport.SportInternalID {

	case 1:

		_sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time, round_name.short_round_name, round_name.round_name, round_name.round_url," +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color, " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" arcache.match_weather, arcache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_aussie_rules_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_aussie_rules_cache AS arcache ON arcache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_aussie_rules_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE " + _searchStr + "" +
			"  matches.status = ? AND matches.league_id = ? " +
			" ORDER BY matches.season_id ASC, matches.round_id ASC, match_date ASC, match_time ASC " + limitstr

	case 7:

		_sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time,  round_name.short_round_name, round_name.round_name, round_name.round_url, " +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color, " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" rlcache.match_weather, rlcache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_rugby_league_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_rugby_league_cache AS rlcache ON rlcache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_rugby_league_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE " + _searchStr + "" +
			"  matches.status = ? AND matches.league_id = ? " +
			" ORDER BY  matches.season_id ASC, matches.round_id ASC, match_date ASC, match_time ASC " + limitstr

	case 10:

		_sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time,  round_name.short_round_name, round_name.round_name, round_name.round_url, " +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color, " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" rucache.match_weather, rucache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_rugby_union_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_rugby_union_cache AS rucache ON rucache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_rugby_union_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE " + _searchStr + "" +
			"  matches.status = ? AND matches.league_id = ? " +
			" ORDER BY  matches.season_id ASC, matches.round_id ASC, match_date ASC, match_time ASC " + limitstr

	}

	return _sqlstr
}

// GetMatchesForGeniusOdds :
func GetMatchesForGeniusOdds(_sqlstr string, objMatchesRecord []isg.GeniusSportsMatch, objLiveOdds []isg.IntMarketInfo, objSport isg.Sport, objLeague isg.League, typeVal string) ([]isg.GeniusSportsMatch, error) {

	rows, err := SportsDb.Query(_sqlstr, "Y", objLeague.LeagueInternalID)
	if err != nil {
		return nil, err
	}
	currentDateTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().In(AEST).Format("2006-01-02 15:04:05"))
	defer rows.Close()
	for rows.Next() {

		objmatch := isg.GeniusSportsMatch{}
		err := rows.Scan(
			&objmatch.MatchID,
			&objmatch.MatchDate,
			&objmatch.MatchTime,
			&objmatch.CounterDate,
			&objmatch.CounterTime,
			&objmatch.Round,
			&objmatch.RoundFullName,
			&objmatch.RoundURL,
			&objmatch.HomeTeamID,
			&objmatch.HomeTeamInternalID,
			&objmatch.HomeTeamName,
			&objmatch.HomeTeamFullName,
			&objmatch.HomeTeamAbbr,
			&objmatch.HomeTeamIcon,
			&objmatch.HomeTeamURL,
			&objmatch.HomeTeamShortName,
			&objmatch.HomeTeamColor,
			&objmatch.AwayTeamID,
			&objmatch.AwayTeamInternalID,
			&objmatch.AwayTeamName,
			&objmatch.AwayTeamFullName,
			&objmatch.AwayTeamAbbr,
			&objmatch.AwayTeamIcon,
			&objmatch.AwayTeamURL,
			&objmatch.AwayTeamShortName,
			&objmatch.AwayTeamColor,
			&objmatch.VenueID,
			&objmatch.VenueName,
			&objmatch.VenueInternalID,
			&objmatch.VenueURL,
			&objmatch.VenueCity,
			&objmatch.VenueCountry,
			&objmatch.MatchStatus,
			&objmatch.MatchWeather,
			&objmatch.MatchDayNight,
			&objmatch.TimeZone,
			&objmatch.MatchReschedule,
		)

		if err != nil {
			return nil, err
		}

		// exclude started match from best matches
		matchDateTime, _ := time.Parse("2006-01-02 15:04:05", objmatch.CounterDate.String+" "+objmatch.CounterTime.String)
		if (typeVal == "best" || typeVal == "upcoming") && matchDateTime.Unix() < currentDateTime.Unix() {
			continue
		}

		objmatch.SportInfo = objSport
		objmatch.LeagueInfo = objLeague

		marketFlag := false
		if typeVal != "market" {

			for _, match := range objLiveOdds {
				if match.MatchID == objmatch.MatchID.Int64 && objmatch.MatchStatus.String == "Y" {
					objmatch.IntMatchOdds = append(objmatch.IntMatchOdds, match)
					marketFlag = true
				}
			}
		} else {
			objmatch.IntMatchOdds = append(objmatch.IntMatchOdds, objLiveOdds...)
			objMatchesRecord = append(objMatchesRecord, objmatch)
			break
		}

		if (typeVal == "upcoming" || typeVal == "best") && marketFlag {
			objMatchesRecord = append(objMatchesRecord, objmatch)
		}
	}

	return objMatchesRecord, nil
}

// GetMatchesGeniusOddsPlunges :
func GetMatchesGeniusOddsPlunges(plungeOddsFluc []isg.GeniusOddsMarket, objSport isg.Sport, leagueID, matchID int) ([]isg.FixtureOdds, error) {

	var liveOdds []isg.FixtureOdds
	var _sqlStr string
	sportID := objSport.SportInternalID
	switch sportID {

	case 1, 7, 10:

		_sqlStr = " SELECT matches.match_id, odds.home_odds, odds.away_odds, first_odds.home_odds, first_odds.away_odds, IFNULL(provider.provider_name,''), " +
			" IFNULL(provider.provider_icon,''), provider.provider_id, matches.home_team_id, matches.away_team_id, odds.plunge_home_odds, odds.plunge_away_odds " +
			" FROM " + objSport.TableNameMatches + " matches" +
			" INNER JOIN " + objSport.TableNameMatches + "_odds odds ON odds.match_id = matches.match_id AND odds.provider_id != ?  " +
			" LEFT JOIN " + objSport.TableNameMatches + "_odds_first AS first_odds ON odds.match_id=first_odds.match_id AND odds.provider_id= first_odds.provider_id " +
			" INNER JOIN isg_providers provider ON provider.provider_id = odds.provider_id " +
			" WHERE matches.match_id = ? " +
			" AND matches.`status` = ? AND matches.league_id = ? AND odds.home_odds IS NOT NULL"
	}

	//fmt.Println(_sqlStr)
	rows, err := SportsDb.Query(_sqlStr, 4, matchID, "Y", leagueID)
	if err != nil {
		return liveOdds, err
	}

	defer rows.Close()
	for rows.Next() {
		var Odds isg.FixtureOdds
		err = rows.Scan(
			&Odds.MatchID,
			&Odds.HomePlunge.NewOdds,
			&Odds.AwayPlunge.NewOdds,
			&Odds.HomePlunge.OpenOdds,
			&Odds.AwayPlunge.OpenOdds,
			&Odds.ProviderInfo.Name,
			&Odds.ProviderInfo.Icon,
			&Odds.ProviderInfo.ProviderId,
			&Odds.HomePlunge.TeamID,
			&Odds.AwayPlunge.TeamID,
			&Odds.HomePlunge.TeamOpenOdds,
			&Odds.AwayPlunge.TeamOpenOdds,
		)

		if err != nil {
			return nil, err
		}

		objHomeFlucs := []*float64{}
		objAwayFlucs := []*float64{}

		/* assigning open odds according to the team past matches basis */
		// home
		if Odds.HomePlunge.TeamOpenOdds != nil {
			Odds.HomePlunge.OpenOdds = Odds.HomePlunge.TeamOpenOdds
		}
		//away
		if Odds.AwayPlunge.TeamOpenOdds != nil {
			Odds.AwayPlunge.OpenOdds = Odds.AwayPlunge.TeamOpenOdds
		}

		if Odds.HomePlunge.NewOdds != nil && Odds.AwayPlunge.NewOdds != nil && Odds.HomePlunge.OpenOdds != nil && Odds.AwayPlunge.OpenOdds != nil {
			//getting change odds
			homechangeOddval, _ := decimal.NewFromFloat((*Odds.HomePlunge.NewOdds - *Odds.HomePlunge.OpenOdds)).Div(decimal.NewFromFloat(*Odds.HomePlunge.OpenOdds - 1)).Mul(decimal.NewFromFloat(float64(100))).Round(2).Float64()
			Odds.HomePlunge.ChangePercentage = &homechangeOddval

			awaychangeOddval, _ := decimal.NewFromFloat((*Odds.AwayPlunge.NewOdds - *Odds.AwayPlunge.OpenOdds)).Div(decimal.NewFromFloat(*Odds.AwayPlunge.OpenOdds - 1)).Mul(decimal.NewFromFloat(float64(100))).Round(2).Float64()
			Odds.AwayPlunge.ChangePercentage = &awaychangeOddval

			// added fluc in plunge
			for _, flucs := range plungeOddsFluc {
				marketPrice := flucs.MarketPrice
				if flucs.MatchID == Odds.MatchID && flucs.ProviderInfo.ProviderId == Odds.ProviderInfo.ProviderId && flucs.MarketTeamID == int64(Odds.HomePlunge.TeamID) {
					objHomeFlucs = append(objHomeFlucs, marketPrice)
				} else if flucs.MatchID == Odds.MatchID && flucs.ProviderInfo.ProviderId == Odds.ProviderInfo.ProviderId && flucs.MarketTeamID == int64(Odds.AwayPlunge.TeamID) {
					objAwayFlucs = append(objAwayFlucs, marketPrice)
				}
			}
			Odds.HomePlunge.Flucs = objHomeFlucs
			Odds.AwayPlunge.Flucs = objAwayFlucs
			liveOdds = append(liveOdds, Odds)
		}

	}
	return liveOdds, nil
}

// GetPlungeMatchesForGeniusOdds :
func GetPlungeMatchesForGeniusOdds(objMatchesRecord []isg.GeniusSportsMatch, livePlungeOdds map[int64][]isg.GeniusOddsPlunge, objSport isg.Sport, objLeague isg.League, typeVal string, matchIDs []string) ([]isg.GeniusSportsMatch, error) {

	var sqlstr, searchStr string
	teamMap := map[int64]int64{}

	matchID := strings.Join(matchIDs, ",")
	searchStr = " AND matches.match_id IN (" + matchID + ") "

	switch objSport.SportInternalID {

	case 1:

		sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time, round_name.short_round_name, round_name.round_name, round_name.round_url, " +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color,  " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" arcache.match_weather, arcache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_aussie_rules_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_aussie_rules_cache AS arcache ON arcache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_aussie_rules_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE matches.status = ? AND matches.league_id = ? " + searchStr +
			" ORDER BY matches.season_id ASC,  matches.round_id ASC, match_date ASC, match_time ASC "

	case 7:

		sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time, round_name.short_round_name, round_name.round_name, round_name.round_url, " +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color, " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" rlcache.match_weather, rlcache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_rugby_league_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_rugby_league_cache AS rlcache ON rlcache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_rugby_league_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE matches.status = ? AND matches.league_id = ? " + searchStr +
			" ORDER BY  matches.season_id ASC, matches.round_id ASC, match_date ASC, match_time ASC  "

	case 10:

		sqlstr = "SELECT matches.match_id, matches.match_date, matches.match_time, counter_date, counter_time, round_name.short_round_name, round_name.round_name, round_name.round_url, " +
			" home.isg_api_id, matches.home_team_id, home.filtername, home.team_name, home.abbreviation, home.icon, home.url, home.short_teamname, home.team_color, " +
			" away.isg_api_id, matches.away_team_id, away.filtername, away.team_name, away.abbreviation, away.icon, away.url, away.short_teamname, away.team_color, " +
			" isg_venue.isg_api_id, isg_venue.filtername, matches.venue_id, isg_venue.friendlyname, isg_venue.city, country.country, matches.status, " +
			" rucache.match_weather, rucache.match_day_night, isg_venue.timezone, matches.is_reschedule " +
			" FROM isg_rugby_union_matches AS matches " +
			" LEFT JOIN isg_team AS home ON home.team_id = matches.home_team_id " +
			" LEFT JOIN isg_team AS away ON away.team_id = matches.away_team_id " +
			" LEFT JOIN isg_venue ON isg_venue.venue_id = matches.venue_id " +
			" LEFT JOIN isg_rugby_union_cache AS rucache ON rucache.match_id = matches.match_id " +
			" LEFT JOIN isg_country country ON country.country_id = isg_venue.country " +
			" LEFT JOIN isg_rugby_union_round round_name ON matches.round_id = round_name.round_id " +
			" WHERE matches.status = ? AND matches.league_id = ? " + searchStr +
			" ORDER BY  matches.season_id ASC, matches.round_id ASC, match_date ASC, match_time ASC  "

	}

	rows, err := SportsDb.Query(sqlstr, "Y", objLeague.LeagueInternalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {

		objmatch := isg.GeniusSportsMatch{}
		err := rows.Scan(
			&objmatch.MatchID,
			&objmatch.MatchDate,
			&objmatch.MatchTime,
			&objmatch.CounterDate,
			&objmatch.CounterTime,
			&objmatch.Round,
			&objmatch.RoundFullName,
			&objmatch.RoundURL,
			&objmatch.HomeTeamID,
			&objmatch.HomeTeamInternalID,
			&objmatch.HomeTeamName,
			&objmatch.HomeTeamFullName,
			&objmatch.HomeTeamAbbr,
			&objmatch.HomeTeamIcon,
			&objmatch.HomeTeamURL,
			&objmatch.HomeTeamShortName,
			&objmatch.HomeTeamColor,
			&objmatch.AwayTeamID,
			&objmatch.AwayTeamInternalID,
			&objmatch.AwayTeamName,
			&objmatch.AwayTeamFullName,
			&objmatch.AwayTeamAbbr,
			&objmatch.AwayTeamIcon,
			&objmatch.AwayTeamURL,
			&objmatch.AwayTeamShortName,
			&objmatch.AwayTeamColor,
			&objmatch.VenueID,
			&objmatch.VenueName,
			&objmatch.VenueInternalID,
			&objmatch.VenueURL,
			&objmatch.VenueCity,
			&objmatch.VenueCountry,
			&objmatch.MatchStatus,
			&objmatch.MatchWeather,
			&objmatch.MatchDayNight,
			&objmatch.TimeZone,
			&objmatch.MatchReschedule,
		)
		if err != nil {
			return nil, err
		}

		objmatch.SportInfo = objSport
		objmatch.LeagueInfo = objLeague

		_, homeok := teamMap[objmatch.HomeTeamInternalID.Int64]
		_, awayok := teamMap[objmatch.AwayTeamInternalID.Int64]

		if plungeOdds, ok := livePlungeOdds[objmatch.MatchID.Int64]; ok && !homeok && !awayok && objmatch.MatchStatus.String == "Y" {

			teamMap[objmatch.HomeTeamInternalID.Int64] = objmatch.HomeTeamInternalID.Int64
			teamMap[objmatch.AwayTeamInternalID.Int64] = objmatch.AwayTeamInternalID.Int64

			objmatch.PlungeOddsList = append(objmatch.PlungeOddsList, plungeOdds...)

			objMatchesRecord = append(objMatchesRecord, objmatch)

		}

	}

	return objMatchesRecord, nil
}

// GetSprotsMarketOdds :
func GetSprotsMarketOdds(matchodd isg.Odds) []isg.MarketOdds {
	objMarketOdds := []isg.MarketOdds{}
	objMarketOdd := isg.MarketOdds{}

	// h2h market
	if matchodd.HomeOdds != nil {
		objMarketOptions := []isg.MarketOptions{}
		objMarketOption := isg.MarketOptions{}
		var objHomeOdds, objAwayOdds, objDrawOdds isg.OddsInfo
		//home
		objHomeOdds.OpenOdds = matchodd.HomeFirstOdds
		if matchodd.HomeFirstOdds == nil {
			objHomeOdds.OpenOdds = matchodd.HomeOdds
		}
		homeoddsfluctuation, _ := decimal.NewFromFloat(*matchodd.HomeOdds).Sub(decimal.NewFromFloat(*objHomeOdds.OpenOdds)).Round(2).Float64()
		objHomeOdds.NewOdds = matchodd.HomeOdds
		objHomeOdds.FlucPer = &homeoddsfluctuation
		objHomeOdds.MarketID = matchodd.HomeH2HMarket
		objMarketOption.HomeMarketOdds = &objHomeOdds
		//away
		objAwayOdds.OpenOdds = matchodd.AwayFirstOdds
		if matchodd.AwayFirstOdds == nil {
			objAwayOdds.OpenOdds = matchodd.AwayOdds
		}
		awayoddsfluctuation, _ := decimal.NewFromFloat(*matchodd.AwayOdds).Sub(decimal.NewFromFloat(*objAwayOdds.OpenOdds)).Round(2).Float64()
		objAwayOdds.NewOdds = matchodd.AwayOdds
		objAwayOdds.FlucPer = &awayoddsfluctuation
		objAwayOdds.MarketID = matchodd.AwayH2HMarket
		objMarketOption.AwayMarketOdds = &objAwayOdds

		//draw
		if matchodd.DrawOdds != nil {
			objDrawOdds.OpenOdds = matchodd.DrawFirstOdds
			drawoddsfluctuation, _ := decimal.NewFromFloat(*matchodd.DrawOdds).Sub(decimal.NewFromFloat(*objAwayOdds.OpenOdds)).Round(2).Float64()
			objDrawOdds.NewOdds = matchodd.DrawOdds
			objDrawOdds.FlucPer = &drawoddsfluctuation
			objDrawOdds.MarketID = matchodd.DrawH2HMarket
			objMarketOption.DrawyMarketOdds = &objDrawOdds
		}
		objMarketOptions = append(objMarketOptions, objMarketOption)

		objMarketOdd.MarketName = "H2H"
		objMarketOdd.MarketOption = objMarketOptions
		objMarketOdds = append(objMarketOdds, objMarketOdd)
	}

	// line market
	if matchodd.HomeLine != nil {
		objMarketOptions := []isg.MarketOptions{}
		objMarketOption := isg.MarketOptions{}
		var objHomeOdds, objAwayOdds isg.OddsInfo
		//home line odds
		objHomeOdds.NewLine = matchodd.HomeLine
		objHomeOdds.MarketID = matchodd.HomeLineMarket
		objMarketOption.HomeMarketOdds = &objHomeOdds
		//away line odds
		objAwayOdds.NewLine = matchodd.AwayLine
		objAwayOdds.MarketID = matchodd.AwayLineMarket
		objMarketOption.AwayMarketOdds = &objAwayOdds
		objMarketOption.MarketType = "line"
		objMarketOptions = append(objMarketOptions, objMarketOption)

		objMarketOption = isg.MarketOptions{}
		objHomeLine, objAwayLine := isg.OddsInfo{}, isg.OddsInfo{}
		//home line
		objHomeLine.NewOdds = matchodd.HomeLineOdds
		objHomeLine.MarketID = matchodd.HomeLineMarket
		objMarketOption.HomeMarketOdds = &objHomeLine
		//away line
		objAwayLine.NewOdds = matchodd.AwayLineOdds
		objAwayLine.MarketID = matchodd.AwayLineMarket
		objMarketOption.AwayMarketOdds = &objAwayLine
		objMarketOption.MarketType = "line odds"
		objMarketOptions = append(objMarketOptions, objMarketOption)

		objMarketOdd.MarketName = "Line"
		objMarketOdd.MarketOption = objMarketOptions
		objMarketOdds = append(objMarketOdds, objMarketOdd)
	}

	// closing market
	if matchodd.ClosingTotal != nil {
		objMarketOptions := []isg.MarketOptions{}
		objMarketOption := isg.MarketOptions{}
		var objClosing, objOver, objUnder isg.OddsInfo
		//closing
		objClosing.NewTotal = matchodd.ClosingTotal
		objMarketOption.ClosingMarketOdds = &objClosing
		//over
		objOver.NewOdds = matchodd.OverOdds
		objOver.MarketID = matchodd.OverMarket
		objMarketOption.OverMarketOdds = &objOver
		//under
		objUnder.NewOdds = matchodd.UnderOdds
		objUnder.MarketID = matchodd.UnderMarket
		objMarketOption.UnderMarketOdds = &objUnder
		objMarketOptions = append(objMarketOptions, objMarketOption)

		objMarketOdd.MarketName = "Closing Total"
		objMarketOdd.MarketOption = objMarketOptions
		objMarketOdds = append(objMarketOdds, objMarketOdd)
	}

	return objMarketOdds
}

// GetMatchesProviderMarketOdds :
func GetMatchesProviderMarketOdds(marketFlucs []isg.GeniusOddsMarket, objSport isg.Sport, leagueID, matchID int, typeVal string) ([]isg.GeniusOddsMarket, error) {
	var liveOdds []isg.GeniusOddsMarket
	var sqlstr, searchStr, search string

	flucMap := map[string]string{}
	sportID := objSport.SportInternalID
	searchStr = "concat(matches.counter_date, ' ', matches.counter_time) BETWEEN '" + time.Now().In(AEST).Format("2006-01-02 15:04:05") + "' AND " +
		" '" + time.Now().In(AEST).AddDate(0, 0, 8).Format("2006-01-02 15:04:05") + "' AND "

	if matchID != 0 {
		searchStr = "matches.match_id = " + strconv.Itoa(matchID) + " AND "
	}

	switch sportID {

	case 1, 7, 10: // AFL
		if typeVal == "best" {
			search = " market.isg_api_id IN ('win', 'loss', 'draw', 'over', 'under','cover') "
		} else if typeVal == "upcoming" {
			search = " market.isg_api_id IN ('win') "
		}
		sqlstr = "SELECT matches.match_id, matches.home_team_id, matches.away_team_id, market.market_id, market.market_name, IFNULL(marketcategory.category_id,0), IFNULL(marketcategory.category_name,''), " +
			" marketodds.team_id, marketodds.market_price, marketodds.market_val, marketodds.provider_market_id, IFNULL(provider.provider_name,''), IFNULL(provider.provider_icon,''), marketodds.provider_id, " +
			" IFNULL(provider.genius_odds_sequence, 0), market.isg_api_id " +
			" FROM " + objSport.TableNameMatches + " AS matches " +
			" INNER JOIN isg_geniusodds_marketodds marketodds ON marketodds.match_id = matches.match_id AND marketodds.sport_id = ? AND marketodds.league_level_id = ? " +
			" AND marketodds.provider_id != ? " +
			" INNER JOIN isg_market market ON market.market_id = marketodds.market_id " +
			" LEFT JOIN isg_market_category marketcategory ON marketcategory.category_id = market.category_id " +
			" LEFT JOIN isg_providers provider ON marketodds.provider_id= provider.provider_id " +
			" WHERE " + searchStr + "" + search +
			" AND matches.status = ? AND matches.league_id = ?  AND marketodds.`status`= ? " +
			" ORDER BY matches.match_id,  market.category_id, market.market_id"
	}

	rows, err := SportsDb.Query(sqlstr, sportID, leagueID, 4, "Y", leagueID, 1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var odds isg.GeniusOddsMarket
		if sportID != 4 {
			err = rows.Scan(
				&odds.MatchID,
				&odds.HomeTeamID,
				&odds.AwayTeamID,
				&odds.MarketID,
				&odds.MarketName,
				&odds.CategoryID,
				&odds.CategoryName,
				&odds.MarketTeamID,
				&odds.MarketPrice,
				&odds.MarketVal,
				&odds.ProviderMarketID,
				&odds.ProviderInfo.Name,
				&odds.ProviderInfo.Icon,
				&odds.ProviderInfo.ProviderId,
				&odds.ProviderInfo.GeniusOddsSequence,
				&odds.ISGapiID,
			)
		}

		if err != nil {
			return nil, err
		}

		objFlucs := []*float64{}
		//get open odds
		for _, flucs := range marketFlucs {
			matchid := strconv.Itoa(int(flucs.MatchID))
			marketid := strconv.Itoa(int(flucs.MarketID))
			teamid := strconv.Itoa(int(flucs.MarketTeamID))
			providerid := flucs.ProviderInfo.ProviderId
			flucStr := matchid + marketid + teamid + providerid

			_, flucok := flucMap[flucStr]

			if !flucok && flucs.MatchID == odds.MatchID && flucs.MarketID == odds.MarketID && flucs.MarketTeamID == odds.MarketTeamID && flucs.ProviderInfo.ProviderId == odds.ProviderInfo.ProviderId {
				odds.MarketFlucPrice = flucs.MarketPrice
				odds.MarketFlucVal = flucs.MarketVal

				matchid = strconv.Itoa(int(flucs.MatchID))
				marketid = strconv.Itoa(int(flucs.MarketID))
				teamid = strconv.Itoa(int(flucs.MarketTeamID))
				providerid = flucs.ProviderInfo.ProviderId
				flucStr = matchid + marketid + teamid + providerid
				flucMap[flucStr] = flucStr
			}

			if flucs.MatchID == odds.MatchID && flucs.MarketID == odds.MarketID && flucs.MarketTeamID == odds.MarketTeamID && flucs.ProviderInfo.ProviderId == odds.ProviderInfo.ProviderId {
				marketPrice := flucs.MarketPrice
				if (odds.CategoryName == "Line" || odds.CategoryName == "Total") && flucs.MarketVal != nil {
					marketPrice = flucs.MarketVal
				}
				objFlucs = append(objFlucs, marketPrice)
			}
		}
		odds.Flucs = objFlucs
		liveOdds = append(liveOdds, odds)
	}

	return liveOdds, nil
}

// GetMarketMatchesProviderOdds :
func GetMarketMatchesProviderOdds(marketFlucs []isg.GeniusOddsMarket, objSport isg.Sport, leagueID, matchID int) ([]isg.GeniusOddsMarket, error) {
	var liveOdds []isg.GeniusOddsMarket
	var sqlstr, searchStr, search string

	flucMap := map[string]string{}

	searchStr = "concat(matches.counter_date, ' ', matches.counter_time) <= '" + time.Now().In(AEST).AddDate(0, 0, 8).Format("2006-01-02 15:04:05") + "'"
	if matchID != 0 {
		searchStr = "matches.match_id = " + strconv.Itoa(matchID)
	}
	sportID := objSport.SportInternalID
	switch sportID {

	case 1, 7, 10: // AFL
		sqlstr = " SELECT matches.match_id, matches.home_team_id, matches.away_team_id, market.market_id, marketodds.team_id, marketodds.market_price, marketodds.market_val, " +
			" marketodds.provider_market_id, IFNULL(provider.provider_name,'') AS provider_name, IFNULL(provider.provider_icon,'') AS provider_icon, provider.provider_id, " +
			" IFNULL(provider.genius_odds_sequence, 0), marketmap.parent_id, " +
			" marketmap.market_name, map.market_name AS category_name, isg_market_category_group.group_name, IFNULL(marketodds.market_display_name,'') AS display_name, map.sequence " +
			" FROM " + objSport.TableNameMatches + " AS matches " +
			" INNER JOIN isg_geniusodds_marketodds marketodds ON marketodds.match_id = matches.match_id AND marketodds.sport_id = ? AND marketodds.league_level_id = ? " +
			" AND marketodds.provider_id != ? " +
			" INNER JOIN isg_market market ON market.market_id = marketodds.market_id " +
			" INNER JOIN isg_geniusodds_markets_mapping AS marketmap ON marketmap.market_id = market.market_id " +
			" LEFT JOIN isg_geniusodds_markets_mapping AS map ON map.mapping_id = marketmap.parent_id " +
			" LEFT JOIN isg_market_category_group ON isg_market_category_group.group_id = map.group_id " +
			" LEFT JOIN isg_providers provider ON marketodds.provider_id= provider.provider_id " +
			" WHERE " + searchStr + "" + search +
			" AND matches.status = ? AND matches.league_id = ? AND marketodds.`status`= ? " +
			" ORDER BY matches.match_id, isg_market_category_group.group_id, marketmap.sequence, market.market_id, marketodds.provider_id, " +
			" IF(matches.home_team_id < matches.away_team_id, team_id, 0) ASC, team_id DESC "
	}
	//fmt.Println(sqlstr)
	rows, err := SportsDb.Query(sqlstr, sportID, leagueID, 4, "Y", leagueID, 1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var odds isg.GeniusOddsMarket
		if sportID != 4 {
			err = rows.Scan(
				&odds.MatchID,
				&odds.HomeTeamID,
				&odds.AwayTeamID,
				&odds.MarketID,
				&odds.MarketTeamID,
				&odds.MarketPrice,
				&odds.MarketVal,
				&odds.ProviderMarketID,
				&odds.ProviderInfo.Name,
				&odds.ProviderInfo.Icon,
				&odds.ProviderInfo.ProviderId,
				&odds.ProviderInfo.GeniusOddsSequence,
				&odds.ParentID,
				&odds.MarketName,
				&odds.CategoryName,
				&odds.GroupName,
				&odds.DisplayName,
				&odds.Sequence,
			)
		}

		if err != nil {
			return nil, err
		}

		var prevLine *float64
		objFlucs := []*float64{}
		//get open odds
		for _, flucs := range marketFlucs {
			matchid := strconv.Itoa(int(flucs.MatchID))
			marketid := strconv.Itoa(int(flucs.MarketID))
			teamid := strconv.Itoa(int(flucs.MarketTeamID))
			providerid := flucs.ProviderInfo.ProviderId
			flucStr := matchid + marketid + teamid + providerid

			_, flucok := flucMap[flucStr]

			if !flucok && flucs.MatchID == odds.MatchID && flucs.MarketID == odds.MarketID && flucs.MarketTeamID == odds.MarketTeamID && flucs.ProviderInfo.ProviderId == odds.ProviderInfo.ProviderId {
				odds.MarketFlucPrice = flucs.MarketPrice
				odds.MarketFlucVal = flucs.MarketVal

				matchid = strconv.Itoa(int(flucs.MatchID))
				marketid = strconv.Itoa(int(flucs.MarketID))
				teamid = strconv.Itoa(int(flucs.MarketTeamID))
				providerid = flucs.ProviderInfo.ProviderId
				flucStr = matchid + marketid + teamid + providerid
				flucMap[flucStr] = flucStr
			}

			if flucs.MatchID == odds.MatchID && flucs.MarketID == odds.MarketID && flucs.MarketTeamID == odds.MarketTeamID && flucs.ProviderInfo.ProviderId == odds.ProviderInfo.ProviderId {
				if flucs.MarketVal != nil {
					if (prevLine == nil) || (*prevLine != *flucs.MarketVal) {
						objFlucs = append(objFlucs, flucs.MarketVal)
						prevLine = flucs.MarketVal
					}
				} else {
					objFlucs = append(objFlucs, flucs.MarketPrice)
				}
			}
		}
		odds.Flucs = objFlucs
		liveOdds = append(liveOdds, odds)
	}

	return liveOdds, nil
}

// GetMatchesProviderMarketFlucs :
func GetMatchesProviderMarketFlucs(objSport isg.Sport, leagueID, matchID int, typeVal string) ([]isg.GeniusOddsMarket, error) {
	var liveFlucOdds []isg.GeniusOddsMarket
	var sqlstr, searchStr, plungeStr string
	searchStr = "concat(matches.counter_date, ' ', matches.counter_time) BETWEEN '" + time.Now().In(AEST).Format("2006-01-02 15:04:05") + "' AND " +
		" '" + time.Now().In(AEST).AddDate(0, 0, 8).Format("2006-01-02 15:04:05") + "' AND "
	sportID := objSport.SportInternalID
	if matchID != 0 {
		searchStr = "matches.match_id = " + strconv.Itoa(matchID) + " AND "
	}

	switch sportID {

	case 1, 7, 10: // AFL

		if typeVal == "plunge" || typeVal == "upcoming" {
			plungeStr = " AND  market.isg_api_id IN ('win') "
		} else if typeVal == "best" {
			plungeStr = " AND market.isg_api_id IN ('win', 'loss', 'draw', 'over', 'under', 'cover') "
		}

		sqlstr = "SELECT oddsfluc.match_id, oddsfluc.market_id, oddsfluc.provider_id, oddsfluc.team_id, oddsfluc.market_price, oddsfluc.market_val, marketcategory.category_name " +
			" FROM " + objSport.TableNameMatches + " AS matches " +
			" INNER JOIN isg_geniusodds_marketodds marketodds ON marketodds.match_id = matches.match_id AND marketodds.sport_id = ? AND marketodds.league_level_id = ? " +
			" AND marketodds.provider_id != ? " +
			" INNER JOIN isg_market market ON market.market_id = marketodds.market_id " +
			" LEFT JOIN isg_market_category marketcategory ON marketcategory.category_id = market.category_id " +
			" LEFT JOIN isg_geniusodds_marketodds_flucs oddsfluc ON oddsfluc.match_id = marketodds.match_id AND oddsfluc.market_id = marketodds.market_id " +
			" AND oddsfluc.team_id = marketodds.team_id AND oddsfluc.provider_id = marketodds.provider_id " +
			" WHERE " + searchStr + "" +
			"  matches.status = ? AND matches.league_id = ? AND marketodds.`status`= ? " + plungeStr +
			" ORDER BY matches.match_id, market.market_id, market.category_id, oddsfluc.last_update "
	}

	rows, err := SportsDb.Query(sqlstr, sportID, leagueID, 4, "Y", leagueID, 1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var flucOdds isg.GeniusOddsMarket
		if sportID != 4 {
			err = rows.Scan(
				&flucOdds.MatchID,
				&flucOdds.MarketID,
				&flucOdds.ProviderInfo.ProviderId,
				&flucOdds.MarketTeamID,
				&flucOdds.MarketPrice,
				&flucOdds.MarketVal,
				&flucOdds.CategoryName,
			)
		}

		if err != nil {
			return nil, err
		}

		liveFlucOdds = append(liveFlucOdds, flucOdds)
	}

	return liveFlucOdds, nil
}

// GetMarketMatchesProviderFlucs :
func GetMarketMatchesProviderFlucs(objSport isg.Sport, leagueID, matchID int, typeVal string) ([]isg.GeniusOddsMarket, error) {
	var liveFlucOdds []isg.GeniusOddsMarket
	var sqlstr, searchStr string

	searchStr = "concat(matches.counter_date, ' ', matches.counter_time) <= '" + time.Now().In(AEST).AddDate(0, 0, 4).Format("2006-01-02 15:04:05") + "'"
	if matchID != 0 {
		searchStr = "matches.match_id = " + strconv.Itoa(matchID)
	}
	sportID := objSport.SportInternalID
	switch sportID {

	case 1, 7, 10: // AFL

		sqlstr = " SELECT oddsfluc.match_id, oddsfluc.market_id, oddsfluc.provider_id, oddsfluc.team_id, oddsfluc.market_price, oddsfluc.market_val " +
			" FROM " + objSport.TableNameMatches + " AS matches " +
			" INNER JOIN isg_geniusodds_marketodds marketodds ON marketodds.match_id = matches.match_id AND marketodds.sport_id = ? AND marketodds.league_level_id = ? " +
			" AND marketodds.provider_id != ? " +
			" INNER JOIN isg_market market ON market.market_id = marketodds.market_id " +
			" INNER JOIN isg_geniusodds_markets_mapping AS marketmap  ON marketmap.market_id = market.market_id " +
			" LEFT JOIN isg_geniusodds_marketodds_flucs oddsfluc ON oddsfluc.match_id = marketodds.match_id AND oddsfluc.market_id = marketodds.market_id " +
			" AND oddsfluc.team_id = marketodds.team_id AND oddsfluc.provider_id = marketodds.provider_id " +
			" WHERE " + searchStr + "" +
			" AND matches.status = ? AND matches.league_id = ? AND marketodds.`status`= ? " +
			" ORDER BY matches.match_id, marketmap.sequence, market.market_id, oddsfluc.last_update  "
	}
	rows, err := SportsDb.Query(sqlstr, sportID, leagueID, 4, "Y", leagueID, 1)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var flucOdds isg.GeniusOddsMarket
		if sportID != 4 {
			err = rows.Scan(
				&flucOdds.MatchID,
				&flucOdds.MarketID,
				&flucOdds.ProviderInfo.ProviderId,
				&flucOdds.MarketTeamID,
				&flucOdds.MarketPrice,
				&flucOdds.MarketVal,
			)
		}

		if err != nil {
			return nil, err
		}

		liveFlucOdds = append(liveFlucOdds, flucOdds)
	}

	return liveFlucOdds, nil
}

//GetGeniusOddBestMatch :
func GetGeniusOddBestMatch(sportID, leagueID int, typeVal string) ([]isg.OddsInfo, error) {
	var geniusMatches []isg.OddsInfo

	sqlStr := "SELECT match_id, sport_id, league_level_id, IFNULL(team_id,0), IFNULL(provider_name,''), IFNULL(provider_icon,''), open_price, price, fluc_percentage " +
		" FROM isg_genius_odds_match WHERE matchtype = ? AND status = ? ORDER BY sport_id, league_level_id ASC "

	rows, err := SportsDb.Query(sqlStr, typeVal, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		geniusMatch := isg.OddsInfo{}
		err = rows.Scan(
			&geniusMatch.MatchID,
			&geniusMatch.SportID,
			&geniusMatch.LeagueID,
			&geniusMatch.TeamID,
			&geniusMatch.Name,
			&geniusMatch.Icon,
			&geniusMatch.OpenOdds,
			&geniusMatch.NewOdds,
			&geniusMatch.FlucPer,
		)
		if err != nil {
			return nil, err
		}

		geniusMatches = append(geniusMatches, geniusMatch)
	}
	return geniusMatches, nil
}

//GetGeniusOddPlungeMatch :
func GetGeniusOddPlungeMatch(sportID, leagueID int, typeVal, matchID string) ([]isg.GeniusOddsPlunge, error) {
	var geniusMatches []isg.GeniusOddsPlunge
	var sportLeagueStr string

	if matchID != "" {
		sportLeagueStr = " AND match_id = " + matchID
	}

	sqlStr := "SELECT match_id, sport_id, league_level_id, IFNULL(team_id,0), IFNULL(provider_name,''), IFNULL(provider_icon,''), open_price, price, fluc_percentage " +
		" FROM isg_genius_odds_match WHERE matchtype = ? AND status = ? " + sportLeagueStr + " ORDER BY sport_id, league_level_id ASC"

	rows, err := SportsDb.Query(sqlStr, typeVal, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		geniusMatch := isg.GeniusOddsPlunge{}
		err = rows.Scan(
			&geniusMatch.MatchID,
			&geniusMatch.SportID,
			&geniusMatch.LeagueID,
			&geniusMatch.Plunge.TeamID,
			&geniusMatch.ProviderInfo.Name,
			&geniusMatch.ProviderInfo.Icon,
			&geniusMatch.Plunge.OpenOdds,
			&geniusMatch.Plunge.NewOdds,
			&geniusMatch.Plunge.ChangePercentage,
		)
		if err != nil {
			return nil, err
		}

		geniusMatches = append(geniusMatches, geniusMatch)
	}
	return geniusMatches, nil
}

// GetMatchPreviewContent :
func GetMatchPreviewContent(objSport isg.Sport, objLeague isg.League, matchID int) (string, error) {

	var _sqlstr string
	var preview string
	switch objSport.SportInternalID {

	case 1:

		_sqlstr = "SELECT description FROM isg_aussie_rules_matches_written_content WHERE match_id = ? AND content_id = ? "

	case 2:

		_sqlstr = "SELECT description FROM isg_nflmatches_written_content WHERE match_id = ? AND content_id = ? "

	case 3:

		_sqlstr = "SELECT description FROM isg_basketball_matches_written_content WHERE match_id = ? AND content_id = ? AND league_id = " + strconv.Itoa(objLeague.LeagueInternalID)

	case 4:

		_sqlstr = "SELECT description FROM isg_soccermatches_written_content WHERE match_id = ? AND content_id = ? "

	case 5:

		_sqlstr = "SELECT description FROM isg_cricket_written_content WHERE match_id = ? AND content_id = ? "

	case 6:

		_sqlstr = "SELECT description FROM isg_tennis_matches_written_content WHERE match_id = ? AND content_id = ? "

	case 7:

		_sqlstr = "SELECT description FROM isg_rugby_league_matches_written_content WHERE match_id = ? AND content_id = ? "

	case 8:

		_sqlstr = "SELECT description FROM isg_hockeymatches_written_content WHERE match_id = ? AND content_id = ? "

	case 9:

		_sqlstr = "SELECT description FROM isg_baseball_matches_written_content WHERE match_id = ? AND content_id = ? "

	case 10:

		_sqlstr = "SELECT description FROM isg_rugby_union_written_content WHERE match_id = ? AND content_id = ? AND league_id = " + strconv.Itoa(objLeague.LeagueInternalID)

	}

	err := SportsDb.QueryRow(_sqlstr, matchID, 1).Scan(&preview)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return preview, nil

}

// GetMatchDetails : Get Match Details as per sport / matchid
func GetMatchDetails(objsport isg.Sport, leagueid, matchid int) (isg.MatchInfo, error) {
	var sqlstr string
	var result isg.MatchInfo

	switch objsport.SportInternalID {
	case 4: // Soccer
		sqlstr = "SELECT match_id, home_team_id, away_team_id " +
			" FROM " + objsport.TableNameMatches + " AS matches " +
			" WHERE match_id = ? AND league_id = " + strconv.Itoa(leagueid)

	}
	err := SportsDb.QueryRow(sqlstr, matchid).Scan(
		&result.MatchID,
		&result.HomeTeamInternalID,
		&result.AwayTeamInternalID,
	)

	if err == sql.ErrNoRows {
		return result, nil
	} else if err != nil {
		return result, err
	}
	return result, nil

}

//UpdateMatchOdds : Update Match Odds as per Sport / Matchid
func UpdateMatchOdds(objsport isg.Sport, homeodds, awayodds, drawodds string, matchid, providerID int) error {
	var sqlstr string
	currentDateTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().In(AEST).Format("2006-01-02 15:04:05"))

	switch objsport.SportInternalID {
	case 4: // Soccer
		sqlstr = "INSERT INTO isg_soccermatches_odds SET match_id = ? , provider_id = ?, home_odds = ?, away_odds = ?, draw_odds = ?, date_added = ?" +
			" ON DUPLICATE KEY UPDATE home_odds = ?, away_odds = ?, draw_odds = ?, date_added = ? "
	}
	stmt, err := SportsDb.Prepare(sqlstr)
	_, err = stmt.Exec(
		matchid,
		providerID,
		homeodds,
		awayodds,
		drawodds,
		currentDateTime,
		homeodds,
		awayodds,
		drawodds,
		currentDateTime,
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateCorrectScoreOdds : Update Soccer Match Correct Odds as per sport / matchid
func UpdateCorrectScoreOdds(correctscoredetails []isg.CorrectScoreDetails, objsport isg.Sport, homeTeamID, awayTeamID string, matchid, providerID int) error {
	var sqlstr string
	var sqlvalues string
	currentDateTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().In(AEST).Format("2006-01-02 15:04:05"))
	updatedcurrentDateTime := currentDateTime.String()
	updatedcurrentDateTime = currentDateTime.Format("2006-01-02 15:04:05")
	fmt.Println(updatedcurrentDateTime)
	switch objsport.SportInternalID {
	case 4: // Soccer
		sqlstr = "INSERT INTO isg_soccer_correct_score_odds (match_id, team_id, provider_id, correct_score, correct_score_odds, correct_score_market, status, date_added) VALUES "

		for _, corectScoreDetail := range correctscoredetails {

			correctscores := strings.Split(corectScoreDetail.CorrectScore, " - ")
			hometeamscore, _ := strconv.Atoi(correctscores[0])
			awayteamscore, _ := strconv.Atoi(correctscores[1])

			teamid := ""
			correctscore := ""
			if hometeamscore > awayteamscore {
				teamid = homeTeamID
				correctscore = strconv.Itoa(hometeamscore) + " - " + strconv.Itoa(awayteamscore)
			} else if awayteamscore > hometeamscore {
				teamid = awayTeamID
				correctscore = strconv.Itoa(awayteamscore) + " - " + strconv.Itoa(hometeamscore)
			} else {
				correctscore = corectScoreDetail.CorrectScore
			}

			correctscoremarket := corectScoreDetail.CorrectScoreMarket
			correctscoreodds := corectScoreDetail.CorrectScoreOdds

			if teamid != "" {
				sqlvalues = sqlvalues + "('" + strconv.Itoa(matchid) + "','" + teamid + "','" + strconv.Itoa(providerID) + "','" + correctscore + "'," +
					correctscoreodds + ",'" + correctscoremarket + "','1','" + updatedcurrentDateTime + "'),"
			} else if teamid == "" {
				sqlvalues = sqlvalues + "('" + strconv.Itoa(matchid) + "','" + homeTeamID + "','" + strconv.Itoa(providerID) + "','" + correctscore + "'," +
					correctscoreodds + ",'" + correctscoremarket + "','1','" + updatedcurrentDateTime + "'),"
				sqlvalues = sqlvalues + "('" + strconv.Itoa(matchid) + "','" + awayTeamID + "','" + strconv.Itoa(providerID) + "','" + correctscore + "'," +
					correctscoreodds + ",'" + correctscoremarket + "','1','" + updatedcurrentDateTime + "'),"
			}

		}

		//trim the last
		sqlstr = sqlstr + sqlvalues[0:len(sqlvalues)-1] + " ON DUPLICATE KEY UPDATE correct_score_market = values(correct_score_market)," +
			" correct_score_odds = values(correct_score_odds), status = 1, date_added = '" + updatedcurrentDateTime + "'"

	}

	stmt, err := SportsDb.Prepare(sqlstr)
	_, err = stmt.Exec()
	defer stmt.Close()
	if err != nil {
		return err

	}

	return nil
}

// GetGeniusMarketMatchID :
func GetGeniusMarketMatchID(objSport isg.Sport, leagueID int, seasonID int, roundWeekDate string, homeTeamID int, awayTeamID int, sortOder string) (int, error) {
	var sqlstr string
	var matchID int

	switch objSport.SportInternalID {
	case 1, 7, 10:
		sqlstr = "SELECT match_id FROM " + objSport.TableNameMatches + " WHERE league_id = ? AND season_id = ? AND round_id = " + roundWeekDate +
			" AND home_team_id = ? AND away_team_id = ? AND status = ?  limit 1"

	case 2, 4:
		sqlstr = "SELECT match_id FROM " + objSport.TableNameMatches + " WHERE league_id = ? AND season_id = ? AND week_id = " + roundWeekDate +
			" AND home_team_id = ? AND away_team_id = ? AND status = ? limit 1"

	case 3, 5, 8, 9:
		if (objSport.SportInternalID == 3 && leagueID == 2) || (objSport.SportInternalID == 5 && leagueID == 1) {
			sqlstr = "SELECT match_id FROM " + objSport.TableNameMatches + " WHERE league_id = ? AND season_id = ? AND round_id = " + roundWeekDate +
				" AND home_team_id = ? AND away_team_id = ? AND status = ?  limit 1"
		} else {
			substr := "ORDER BY match_time 	" + sortOder
			sqlstr = "SELECT match_id FROM " + objSport.TableNameMatches + " WHERE league_id = ? AND season_id = ? AND match_date LIKE '%" + roundWeekDate + "' " +
				" AND home_team_id = ? AND away_team_id = ? AND status = ?  " + substr + " limit 1"
		}
	}

	err := SportsDb.QueryRow(sqlstr, leagueID, seasonID, homeTeamID, awayTeamID, "Y").Scan(&matchID)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return matchID, nil
}
