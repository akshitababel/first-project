package isg

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// BindingGeniusOddsMatches :
func BindingGeniusOddsMatches(objMatches []GeniusSportsMatch, typeVal string) GeniusOddsSportMatch {

	var objSportMatch GeniusOddsSportMatch
	var objGeniusLeague GeniusOddsMatch
	var objLeagueMatch GeniusOddsMatchLeague

	var objMarketOdds []GeniusMarketOdds
	var objMarketOdd GeniusMarketOdds

	//fmt.Println(len(objMatches))
	for i, objmatch := range objMatches {
		var objMarketOptions []GeniusMarketOptions
		var objMarketOption GeniusMarketOptions

		var matchesInfo GeniusOddsMatchesInfo
		matchesInfo.MatchID = objmatch.MatchID.Int64
		matchesInfo.LocalDate = objmatch.MatchDate.String
		matchesInfo.LocalTime = objmatch.MatchTime.String
		matchesInfo.MatchDate = objmatch.CounterDate.String
		matchesInfo.MatchTime = objmatch.CounterTime.String
		if objmatch.TimeZone.Valid {
			matchesInfo.TimeZone = objmatch.TimeZone.String
		}

		matchesInfo.Weather = objmatch.MatchWeather.String
		matchesInfo.Status = "not_started"
		matchesInfo.DayNight = objmatch.MatchDayNight.String
		matchesInfo.IsReschedule = objmatch.MatchReschedule

		// Assign the round/week information
		var objround SportRound
		objround.RoundCode = objmatch.Round.String
		objround.Name = objmatch.RoundFullName.String
		objround.URL = objmatch.RoundURL.String

		// Assign the home team information
		var objhometeam Team
		objhometeam.TeamName = objmatch.HomeTeamName.String
		objhometeam.TeamFlag = objmatch.HomeTeamIcon.String
		objhometeam.Abbreviation = objmatch.HomeTeamAbbr.String
		objhometeam.TeamID = objmatch.HomeTeamID.String
		objhometeam.FullName = objmatch.HomeTeamFullName.String
		objhometeam.TeamColor = objmatch.HomeTeamColor.String

		// Assign the away team information
		var objawayteam Team
		objawayteam.TeamName = objmatch.AwayTeamName.String
		objawayteam.TeamFlag = objmatch.AwayTeamIcon.String
		objawayteam.Abbreviation = objmatch.AwayTeamAbbr.String
		objawayteam.TeamID = objmatch.AwayTeamID.String
		objawayteam.FullName = objmatch.AwayTeamFullName.String
		objawayteam.TeamColor = objmatch.AwayTeamColor.String

		// Assign the venue information
		var objvenue Venue
		objvenue.VenueID = objmatch.VenueID.String
		objvenue.VenueFilterName = objmatch.VenueName.String
		objvenue.VenueCountry = objmatch.VenueCountry.String
		objvenue.VenueCity = objmatch.VenueCity.String

		matchesInfo.Round = objround
		matchesInfo.HomeTeamInfo = objhometeam
		matchesInfo.AwayTeamInfo = objawayteam
		matchesInfo.VenueInfo = objvenue

		// for other markets
		if len(objmatch.IntMatchOdds) > 0 {

			var homeOddsInfo, awayOddsInfo, anyOddsInfo []OddsInfo

			for j, intMarket := range objmatch.IntMatchOdds {

				if len(intMarket.HomeMarketOdds) > 0 {
					for k, marketOdd := range intMarket.HomeMarketOdds {

						var homeOdd, awayOdd OddsInfo
						var objhome, objaway MarketOddsList

						homeOdd.MarketID = marketOdd.MarketID
						homeOdd.OpenOdds = marketOdd.OpenOdds
						homeOdd.NewOdds = marketOdd.NewOdds
						homeOdd.Flucs = marketOdd.Flucs
						if intMarket.CategoryName == "Line" {
							homeOdd.OpenLine = marketOdd.OpenLine
							homeOdd.NewLine = marketOdd.NewLine
						}
						homeOdd.Name = marketOdd.Name
						homeOdd.Icon = marketOdd.Icon
						homeOddsInfo = append(homeOddsInfo, homeOdd)

						if len(objmatch.IntMatchOdds[j].AwayMarketOdds) > 0 && k <= len(objmatch.IntMatchOdds[j].AwayMarketOdds)-1 {
							awayOdd.MarketID = objmatch.IntMatchOdds[j].AwayMarketOdds[k].MarketID
							awayOdd.OpenOdds = objmatch.IntMatchOdds[j].AwayMarketOdds[k].OpenOdds
							awayOdd.NewOdds = objmatch.IntMatchOdds[j].AwayMarketOdds[k].NewOdds
							awayOdd.Flucs = objmatch.IntMatchOdds[j].AwayMarketOdds[k].Flucs
							if intMarket.CategoryName == "Line" {
								awayOdd.OpenLine = objmatch.IntMatchOdds[j].AwayMarketOdds[k].OpenLine
								awayOdd.NewLine = objmatch.IntMatchOdds[j].AwayMarketOdds[k].NewLine
							}
							awayOdd.Name = objmatch.IntMatchOdds[j].AwayMarketOdds[k].Name
							awayOdd.Icon = objmatch.IntMatchOdds[j].AwayMarketOdds[k].Icon
							awayOddsInfo = append(awayOddsInfo, awayOdd)
						}

						if (len(intMarket.HomeMarketOdds)-1) == k || intMarket.HomeMarketOdds[k+1].MarketInternalID != marketOdd.MarketInternalID {
							if typeVal != "market" {
								if len(homeOddsInfo) > 0 {
									objhome.ProviderList = homeOddsInfo[:1]
								}
								if len(awayOddsInfo) > 0 {
									objaway.ProviderList = awayOddsInfo[:1]
								}
							} else {
								objhome.ProviderList = homeOddsInfo
								objaway.ProviderList = awayOddsInfo
							}

							objMarketOption.MarketType = marketOdd.MarketName
							objMarketOption.GeniusHomeMarketOdds = &objhome
							objMarketOption.GeniusAwayMarketOdds = &objaway
							objMarketOptions = append(objMarketOptions, objMarketOption)
							homeOddsInfo, awayOddsInfo = []OddsInfo{}, []OddsInfo{}
						}
					}
				}

				//others
				if len(intMarket.AnyMarketOdds) > 0 {
					for k, marketOdd := range intMarket.AnyMarketOdds {
						var objMarketOption GeniusMarketOptions
						var anyOdd OddsInfo
						var objany MarketOddsList

						anyOdd.MarketID = marketOdd.MarketID
						anyOdd.OpenOdds = marketOdd.OpenOdds
						anyOdd.NewOdds = marketOdd.NewOdds
						if intMarket.CategoryName == "Total" {
							anyOdd.OpenTotal = marketOdd.OpenTotal
							anyOdd.NewTotal = marketOdd.NewTotal
						}
						anyOdd.Flucs = marketOdd.Flucs
						anyOdd.Name = marketOdd.Name
						anyOdd.Icon = marketOdd.Icon
						anyOddsInfo = append(anyOddsInfo, anyOdd)

						if (len(intMarket.AnyMarketOdds)-1) == k || intMarket.AnyMarketOdds[k+1].MarketInternalID != marketOdd.MarketInternalID {

							if typeVal != "market" {
								if len(anyOddsInfo) > 0 {
									objany.ProviderList = anyOddsInfo[:1]
								}
							} else {
								objany.ProviderList = anyOddsInfo
							}
							objMarketOption.MarketType = marketOdd.MarketName
							objMarketOption.GeniusAnyMarketOdds = &objany
							objMarketOptions = append(objMarketOptions, objMarketOption)
							anyOddsInfo = []OddsInfo{}
						}
					}

				}

				objMarketOdd.MarketName = intMarket.CategoryName
				objMarketOdd.MarketOption = objMarketOptions
				objMarketOdds = append(objMarketOdds, objMarketOdd)

				objMarketOptions = []GeniusMarketOptions{}
				objMarketOdd = GeniusMarketOdds{}

			}
			matchesInfo.Market = objMarketOdds
			objMarketOdds = []GeniusMarketOdds{}
		}

		/*------------------------------------PLUNGE----------------------------------------*/

		var objPlungeList MarketOddsList
		var providerOrder int = 1

		for p, plunge := range objmatch.PlungeOddsList {

			if p == 0 {

				if plunge.Plunge.TeamID == int(objmatch.HomeTeamInternalID.Int64) || plunge.Plunge.OddsType == "home" {
					objPlungeList.TeamInfo = &objhometeam
				} else if plunge.Plunge.TeamID == int(objmatch.AwayTeamInternalID.Int64) || plunge.Plunge.OddsType == "away" {
					objPlungeList.TeamInfo = &objawayteam
				}
			}

			var objPlunges OddsInfo

			objPlunges.Icon = plunge.ProviderInfo.Icon
			objPlunges.Name = plunge.ProviderInfo.Name
			objPlunges.ProviderOrder = providerOrder
			objPlunges.OpenOdds = plunge.Plunge.OpenOdds
			objPlunges.NewOdds = plunge.Plunge.NewOdds
			objPlunges.FlucPer = plunge.Plunge.ChangePercentage
			objPlunges.Flucs = plunge.Plunge.Flucs

			providerOrder++
			objPlungeList.ProviderList = append(objPlungeList.ProviderList, objPlunges)

		}

		if len(objPlungeList.ProviderList) > 0 {
			matchesInfo.GeniusOddsPlung = &objPlungeList
		}
		/*------------------------------------//PLUNGE----------------------------------------*/

		if (len(objMatches)-1) == i || objmatch.MatchID != objMatches[i+1].MatchID {
			objLeagueMatch.Matches = append(objLeagueMatch.Matches, matchesInfo)
		}

		if (len(objMatches)-1) != i && (objmatch.LeagueInfo.LeagueInternalID != objMatches[i+1].LeagueInfo.LeagueInternalID ||
			objmatch.SportInfo.SportInternalID != objMatches[i+1].SportInfo.SportInternalID) {

			leagues := strings.Split(objmatch.LeagueInfo.LeagueName, " - ")
			objLeagueMatch.Leaguename = leagues[1]
			objLeagueMatch.LeagueURL = objmatch.LeagueInfo.LeagueEntityKey
			objGeniusLeague.Leagues = append(objGeniusLeague.Leagues, objLeagueMatch)

			objLeagueMatch = GeniusOddsMatchLeague{}

			if objmatch.SportInfo.SportInternalID != objMatches[i+1].SportInfo.SportInternalID {

				objGeniusLeague.SportID = objmatch.SportInfo.SportAPICode
				objGeniusLeague.SportName = objmatch.SportInfo.SportName
				objGeniusLeague.SportURL = objmatch.SportInfo.SportURL
				objSportMatch.Sport = append(objSportMatch.Sport, objGeniusLeague)
				objGeniusLeague = GeniusOddsMatch{}
			}

		}
		if (len(objMatches) - 1) == i {

			leagues := strings.Split(objmatch.LeagueInfo.LeagueName, " - ")
			objLeagueMatch.Leaguename = leagues[1]
			objLeagueMatch.LeagueURL = objmatch.LeagueInfo.LeagueEntityKey
			objGeniusLeague.Leagues = append(objGeniusLeague.Leagues, objLeagueMatch)
			objGeniusLeague.SportID = objmatch.SportInfo.SportAPICode
			objGeniusLeague.SportName = objmatch.SportInfo.SportName
			objGeniusLeague.SportURL = objmatch.SportInfo.SportURL
			objSportMatch.Sport = append(objSportMatch.Sport, objGeniusLeague)
			objLeagueMatch = GeniusOddsMatchLeague{}
			objGeniusLeague = GeniusOddsMatch{}

		}
	}
	return objSportMatch
}

// BindingGeniusOddsMarketMatches :
func BindingGeniusOddsMarketMatches(objMatches []GeniusSportsMatch, typeVal string, plungeMatch []GeniusOddsPlunge) GeniusOddsSportMatch {

	var objSportMatch GeniusOddsSportMatch
	var objGeniusLeague GeniusOddsMatch
	var objLeagueMatch GeniusOddsMatchLeague

	var objMarketGroup []GeniusGroups

	var objMarketOdds []GeniusMarketOdds
	var objMarketOdd GeniusMarketOdds

	//fmt.Println(len(objMatches))
	for i, objmatch := range objMatches {
		var objMarketOptions []MarketOddsList
		//var objMarketOption MarketOddsList

		var matchesInfo GeniusOddsMatchesInfo
		matchesInfo.MatchID = objmatch.MatchID.Int64
		matchesInfo.LocalDate = objmatch.MatchDate.String
		matchesInfo.LocalTime = objmatch.MatchTime.String
		matchesInfo.MatchDate = objmatch.CounterDate.String
		matchesInfo.MatchTime = objmatch.CounterTime.String
		if objmatch.TimeZone.Valid {
			matchesInfo.TimeZone = objmatch.TimeZone.String
		}

		matchesInfo.Weather = objmatch.MatchWeather.String
		matchesInfo.Status = "not_started"
		matchesInfo.DayNight = objmatch.MatchDayNight.String
		matchesInfo.IsReschedule = objmatch.MatchReschedule

		// Assign the round information
		var objround SportRound
		objround.RoundCode = objmatch.Round.String
		objround.Name = objmatch.RoundFullName.String
		objround.URL = objmatch.RoundURL.String

		// Assign the home team information
		var objhometeam Team
		objhometeam.TeamName = objmatch.HomeTeamName.String
		objhometeam.TeamFlag = objmatch.HomeTeamIcon.String
		objhometeam.Abbreviation = objmatch.HomeTeamAbbr.String
		objhometeam.TeamID = objmatch.HomeTeamID.String
		objhometeam.FullName = objmatch.HomeTeamFullName.String
		objhometeam.TeamColor = objmatch.HomeTeamColor.String

		// Assign the away team information
		var objawayteam Team
		objawayteam.TeamName = objmatch.AwayTeamName.String
		objawayteam.TeamFlag = objmatch.AwayTeamIcon.String
		objawayteam.Abbreviation = objmatch.AwayTeamAbbr.String
		objawayteam.TeamID = objmatch.AwayTeamID.String
		objawayteam.FullName = objmatch.AwayTeamFullName.String
		objawayteam.TeamColor = objmatch.AwayTeamColor.String

		// Assign the venue information
		var objvenue Venue
		objvenue.VenueID = objmatch.VenueID.String
		objvenue.VenueFilterName = objmatch.VenueName.String
		objvenue.VenueCountry = objmatch.VenueCountry.String
		objvenue.VenueCity = objmatch.VenueCity.String

		matchesInfo.Round = objround
		matchesInfo.HomeTeamInfo = objhometeam
		matchesInfo.AwayTeamInfo = objawayteam
		matchesInfo.VenueInfo = objvenue
		if len(plungeMatch) > 0 {
			if int(objmatch.HomeTeamInternalID.Int64) == plungeMatch[0].Plunge.TeamID {
				matchesInfo.IsPlunge = objmatch.HomeTeamID.String
			} else if int(objmatch.AwayTeamInternalID.Int64) == plungeMatch[0].Plunge.TeamID {
				matchesInfo.IsPlunge = objmatch.AwayTeamID.String
			}
		}

		// for other markets

		if len(objmatch.IntMatchOdds) > 0 {

			var anyOddsInfo []OddsInfo

			for j, intMarket := range objmatch.IntMatchOdds {
				var MarketName string
				if len(intMarket.AnyMarketOdds) > 0 {

					for k, marketOdd := range intMarket.AnyMarketOdds {

						var anyOdd OddsInfo
						var objany MarketOddsList

						anyOdd.MarketID = marketOdd.MarketID
						anyOdd.OpenOdds = marketOdd.OpenOdds
						anyOdd.NewOdds = marketOdd.NewOdds
						anyOdd.Flucs = marketOdd.Flucs

						if strings.Contains(intMarket.CategoryName, "Line") {
							anyOdd.OpenLine = marketOdd.OpenLine
							anyOdd.NewLine = marketOdd.NewLine
						} else if strings.Contains(intMarket.CategoryName, "Total") {
							anyOdd.OpenTotal = marketOdd.OpenTotal
							anyOdd.NewTotal = marketOdd.NewTotal
						}

						anyOdd.Name = marketOdd.Name
						anyOdd.Icon = marketOdd.Icon
						anyOdd.ProviderSequence = marketOdd.ProviderSequence
						anyOddsInfo = append(anyOddsInfo, anyOdd)
						objany.MarketType = marketOdd.DisplayName

						objany.AbbrName = strings.Replace(strings.Replace(marketOdd.DisplayName, objmatch.HomeTeamName.String, objmatch.HomeTeamAbbr.String, -1),
							objmatch.AwayTeamName.String, objmatch.AwayTeamAbbr.String, -1)

						if len(intMarket.AnyMarketOdds)-1 == k || marketOdd.DisplayName != intMarket.AnyMarketOdds[k+1].DisplayName {
							objany.ProviderList = anyOddsInfo
							//objMarketOption.GeniusAnyMarketOdds = &objany
							objMarketOptions = append(objMarketOptions, objany)
							anyOddsInfo = []OddsInfo{}
							MarketName = marketOdd.MarketName
						}
					}
				}

				// Sort the fgs according to thier best provider new odds
				if MarketName == "First Goal Scorer" {
					sort.Sort(FGSSort(objMarketOptions))
				}

				objMarketOdd.MarketName = intMarket.CategoryName
				objMarketOdd.MatchMarketOption = objMarketOptions
				objMarketOdds = append(objMarketOdds, objMarketOdd)

				objMarketOptions = []MarketOddsList{}

				if (len(objmatch.IntMatchOdds)-1) == j || intMarket.GroupName != objmatch.IntMatchOdds[j+1].GroupName {
					var objGroup GeniusGroups
					objGroup.GroupName = intMarket.GroupName
					objGroup.Market = objMarketOdds
					objMarketGroup = append(objMarketGroup, objGroup)
					objMarketOdds = []GeniusMarketOdds{}
					objMarketOdd = GeniusMarketOdds{}

				}

			}
			matchesInfo.Groups = objMarketGroup
			objMarketOdds = []GeniusMarketOdds{}
			objMarketGroup = []GeniusGroups{}
		}

		if (len(objMatches)-1) == i || objmatch.MatchID != objMatches[i+1].MatchID {
			objLeagueMatch.Matches = append(objLeagueMatch.Matches, matchesInfo)
		}

		if (len(objMatches)-1) != i && (objmatch.LeagueInfo.LeagueInternalID != objMatches[i+1].LeagueInfo.LeagueInternalID ||
			objmatch.SportInfo.SportInternalID != objMatches[i+1].SportInfo.SportInternalID) {

			leagues := strings.Split(objmatch.LeagueInfo.LeagueName, " - ")
			objLeagueMatch.Leaguename = leagues[1]
			objLeagueMatch.LeagueURL = objmatch.LeagueInfo.LeagueEntityKey
			objGeniusLeague.Leagues = append(objGeniusLeague.Leagues, objLeagueMatch)

			objLeagueMatch = GeniusOddsMatchLeague{}

			if objmatch.SportInfo.SportInternalID != objMatches[i+1].SportInfo.SportInternalID {

				objGeniusLeague.SportID = objmatch.SportInfo.SportAPICode
				objGeniusLeague.SportName = objmatch.SportInfo.SportName
				objGeniusLeague.SportURL = objmatch.SportInfo.SportURL
				objSportMatch.Sport = append(objSportMatch.Sport, objGeniusLeague)
				objGeniusLeague = GeniusOddsMatch{}
			}

		}
		if (len(objMatches) - 1) == i {

			leagues := strings.Split(objmatch.LeagueInfo.LeagueName, " - ")
			objLeagueMatch.Leaguename = leagues[1]
			objLeagueMatch.LeagueURL = objmatch.LeagueInfo.LeagueEntityKey
			objGeniusLeague.Leagues = append(objGeniusLeague.Leagues, objLeagueMatch)
			objGeniusLeague.SportID = objmatch.SportInfo.SportAPICode
			objGeniusLeague.SportName = objmatch.SportInfo.SportName
			objGeniusLeague.SportURL = objmatch.SportInfo.SportURL
			objSportMatch.Sport = append(objSportMatch.Sport, objGeniusLeague)
			objLeagueMatch = GeniusOddsMatchLeague{}
			objGeniusLeague = GeniusOddsMatch{}

		}
	}
	return objSportMatch
}

// SorttypeVal :
var SorttypeVal string

// GeniusSortMatchesISG :
type GeniusSortMatchesISG []GeniusSportsMatch

func (a GeniusSortMatchesISG) Len() int      { return len(a) }
func (a GeniusSortMatchesISG) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a GeniusSortMatchesISG) Less(i, j int) bool {

	if a[0].TypeVal != "" {
		SorttypeVal = a[0].TypeVal
	}
	isLess := sortingGeniusMatchesISG(SorttypeVal, a[i], a[j])
	return isLess
}

// sortingGeniusMatchesISG :
func sortingGeniusMatchesISG(typeVal string, objMatchs, objMatchs1 GeniusSportsMatch) bool {

	sequence := SportsSequence[objMatchs.SportInfo.SportInternalID]
	sequence1 := SportsSequence[objMatchs1.SportInfo.SportInternalID]
	
	if typeVal == "upcoming" {

		if objMatchs.CounterDate.String < objMatchs1.CounterDate.String {
			return true
		} else if objMatchs.CounterDate.String == objMatchs1.CounterDate.String && objMatchs.CounterTime.String < objMatchs1.CounterTime.String {
			return true
		} else if objMatchs.CounterDate.String == objMatchs1.CounterDate.String && objMatchs.CounterTime.String == objMatchs1.CounterTime.String && sequence < sequence1 {
			return true
		}
	} else if typeVal == "plunge" {
		if objMatchs.PlungeOddsList != nil && objMatchs1.PlungeOddsList != nil {
			if math.Abs(*objMatchs.PlungeOddsList[0].Plunge.ChangePercentage) > math.Abs(*objMatchs1.PlungeOddsList[0].Plunge.ChangePercentage) {
				return true
			} else if math.Abs(*objMatchs.PlungeOddsList[0].Plunge.ChangePercentage) == math.Abs(*objMatchs1.PlungeOddsList[0].Plunge.ChangePercentage) && sequence < sequence1 {
				return true
			}
		}
	} else if typeVal == "best" {

		if objMatchs.IntMatchOdds[0].HomeMarketOdds[0].NewOdds != nil && objMatchs1.IntMatchOdds[0].HomeMarketOdds[0].NewOdds != nil &&
			(math.Abs(*objMatchs.IntMatchOdds[0].HomeMarketOdds[0].NewOdds-*objMatchs.IntMatchOdds[0].AwayMarketOdds[0].NewOdds) < math.Abs(*objMatchs1.IntMatchOdds[0].HomeMarketOdds[0].NewOdds-*objMatchs1.IntMatchOdds[0].AwayMarketOdds[0].NewOdds)) {
			return true
		} else if objMatchs.IntMatchOdds[0].HomeMarketOdds[0].NewOdds == nil && objMatchs1.IntMatchOdds[0].HomeMarketOdds[0].NewOdds != nil {
			return true
		} else if objMatchs.IntMatchOdds[0].HomeMarketOdds[0].NewOdds != nil && objMatchs1.IntMatchOdds[0].HomeMarketOdds[0].NewOdds != nil &&
			(math.Abs(*objMatchs.IntMatchOdds[0].HomeMarketOdds[0].NewOdds-*objMatchs.IntMatchOdds[0].AwayMarketOdds[0].NewOdds) == math.Abs(*objMatchs1.IntMatchOdds[0].HomeMarketOdds[0].NewOdds-*objMatchs1.IntMatchOdds[0].AwayMarketOdds[0].NewOdds)) &&
			sequence < sequence1 {
			return true
		}

	} 

	return false
}

// PlungeSort :
type PlungeSort []GeniusOddsPlunge

func (a PlungeSort) Len() int      { return len(a) }
func (a PlungeSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a PlungeSort) Less(i, j int) bool {

	return math.Abs(*a[i].Plunge.ChangePercentage) > math.Abs(*a[j].Plunge.ChangePercentage)
}

// FGSSort :
type FGSSort []MarketOddsList

func (a FGSSort) Len() int      { return len(a) }
func (a FGSSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a FGSSort) Less(i, j int) bool {
	if math.Abs(*a[i].ProviderList[0].NewOdds) < math.Abs(*a[j].ProviderList[0].NewOdds) {
		return true
	} else if math.Abs(*a[i].ProviderList[0].NewOdds) == math.Abs(*a[j].ProviderList[0].NewOdds) && a[i].ProviderList[0].ProviderSequence < a[j].ProviderList[0].ProviderSequence {
		return true
	} else if math.Abs(*a[i].ProviderList[0].NewOdds) == math.Abs(*a[j].ProviderList[0].NewOdds) && a[i].ProviderList[0].ProviderSequence == a[j].ProviderList[0].ProviderSequence &&
		a[i].MarketType < a[j].MarketType {
		return true
	}
	return false
}

// MatchMarketSort :
type MatchMarketSort []OddsInfo

func (a MatchMarketSort) Len() int      { return len(a) }
func (a MatchMarketSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a MatchMarketSort) Less(i, j int) bool {
	return sortingForOdds(a, i, j)
}

// sortingForOdds :
func sortingForOdds(a MatchMarketSort, i int, j int) bool {

	if strings.Contains(a[i].CategoryName, "Line") && strings.Contains(a[j].CategoryName, "Line") { // For Line
		if a[i].NewLine != nil && a[j].NewLine != nil && *a[i].NewLine > *a[j].NewLine {
			return true
		} else if a[i].NewLine != nil && a[j].NewLine != nil && *a[i].NewLine == *a[j].NewLine && a[i].ProviderSequence < a[j].ProviderSequence {
			return true
		}
	} else if a[i].MarketName == "Over" && a[j].MarketName == "Over" { // For Total Over

		if a[i].TypeVal == "upcoming" {
			if a[i].NewTotal != nil && a[j].NewTotal != nil && math.Abs(*a[i].NewTotal) > math.Abs(*a[j].NewTotal) {
				return true
			} else if a[i].NewTotal != nil && a[j].NewTotal != nil && (math.Abs(*a[i].NewTotal) == math.Abs(*a[j].NewTotal)) && a[i].ProviderSequence < a[j].ProviderSequence {
				return true
			}
		} else {
			if a[i].NewTotal != nil && a[j].NewTotal != nil && math.Abs(*a[i].NewTotal) < math.Abs(*a[j].NewTotal) {
				return true
			} else if a[i].NewTotal != nil && a[j].NewTotal != nil && (math.Abs(*a[i].NewTotal) == math.Abs(*a[j].NewTotal)) && a[i].ProviderSequence < a[j].ProviderSequence {
				return true
			}
		}

	} else if a[i].MarketName == "Under" && a[j].MarketName == "Under" { // For Total Under
		if a[i].NewTotal != nil && a[j].NewTotal != nil && math.Abs(*a[i].NewTotal) > math.Abs(*a[j].NewTotal) {
			return true
		} else if a[i].NewTotal != nil && a[j].NewTotal != nil && (math.Abs(*a[i].NewTotal) == math.Abs(*a[j].NewTotal)) && a[i].ProviderSequence < a[j].ProviderSequence {
			return true
		}
	} else {
		if a[i].NewOdds != nil && a[j].NewOdds != nil && math.Abs(*a[i].NewOdds) > math.Abs(*a[j].NewOdds) {
			return true
		} else if a[i].NewOdds != nil && a[j].NewOdds != nil && (math.Abs(*a[i].NewOdds) == math.Abs(*a[j].NewOdds)) && a[i].ProviderSequence < a[j].ProviderSequence {
			return true
		}
	}

	return false
}

//MakingGeniusLiveOddsSort :
func MakingGeniusLiveOddsSort(objLiveOdds []GeniusOddsMarket, sportID int, typeVal string) []IntMarketInfo {

	var geniusMarketOdds []IntMarketInfo
	var geniusMarket IntMarketInfo

	var objHomeMarketOdds, objAwayMarketOdds, objAnyMarketOdds []OddsInfo

	for i, objbest := range objLiveOdds {

		var objOdd OddsInfo

		// for other markets
		if objbest.MarketPrice != nil {
			objOdd.Name = objbest.ProviderInfo.Name
			objOdd.Icon = objbest.ProviderInfo.Icon
			objOdd.ProviderSequence = objbest.ProviderInfo.GeniusOddsSequence
			objOdd.TypeVal = typeVal
			if objbest.HomeTeamID == objbest.MarketTeamID {
				objOdd.MarketID = objbest.ProviderMarketID
				objOdd.OpenOdds = objbest.MarketFlucPrice
				objOdd.NewOdds = objbest.MarketPrice
				objOdd.MarketInternalID = int(objbest.MarketID)
				objOdd.MarketName = objbest.MarketName
				objOdd.Flucs = objbest.Flucs
				//condition for line
				if objbest.CategoryName == "Line" && objbest.MarketVal != nil {
					objOdd.OpenLine = objbest.MarketFlucVal
					objOdd.NewLine = objbest.MarketVal
				}
				objHomeMarketOdds = append(objHomeMarketOdds, objOdd)
			} else if objbest.AwayTeamID == objbest.MarketTeamID {
				objOdd.MarketID = objbest.ProviderMarketID
				objOdd.OpenOdds = objbest.MarketFlucPrice
				objOdd.NewOdds = objbest.MarketPrice
				objOdd.MarketInternalID = int(objbest.MarketID)
				objOdd.MarketName = objbest.MarketName
				objOdd.Flucs = objbest.Flucs
				//condition for line
				if objbest.CategoryName == "Line" && objbest.MarketVal != nil {
					objOdd.OpenLine = objbest.MarketFlucVal
					objOdd.NewLine = objbest.MarketVal
				}
				objAwayMarketOdds = append(objAwayMarketOdds, objOdd)
			} else if objbest.MarketTeamID == 0 {
				objOdd.MarketID = objbest.ProviderMarketID
				objOdd.OpenOdds = objbest.MarketFlucPrice
				objOdd.NewOdds = objbest.MarketPrice
				objOdd.MarketInternalID = int(objbest.MarketID)
				objOdd.MarketName = objbest.MarketName
				objOdd.Flucs = objbest.Flucs
				//condition for closing total
				if objbest.CategoryName == "Total" && objbest.MarketVal != nil {
					objOdd.OpenTotal = objbest.MarketFlucVal
					objOdd.NewTotal = objbest.MarketVal
				}
				objAnyMarketOdds = append(objAnyMarketOdds, objOdd)
			}
		}

		if len(objLiveOdds)-1 == i || objbest.MatchID != objLiveOdds[i+1].MatchID || objbest.MarketID != objLiveOdds[i+1].MarketID {
			sort.Sort(MatchMarketSort(objHomeMarketOdds))
			sort.Sort(MatchMarketSort(objAwayMarketOdds))
			sort.Sort(MatchMarketSort(objAnyMarketOdds))
			geniusMarket.HomeMarketOdds = append(geniusMarket.HomeMarketOdds, objHomeMarketOdds...)
			geniusMarket.AwayMarketOdds = append(geniusMarket.AwayMarketOdds, objAwayMarketOdds...)
			geniusMarket.AnyMarketOdds = append(geniusMarket.AnyMarketOdds, objAnyMarketOdds...)
			objHomeMarketOdds, objAwayMarketOdds, objAnyMarketOdds = []OddsInfo{}, []OddsInfo{}, []OddsInfo{}
		}

		// here we added some static condition to maintain market category node
		if len(objLiveOdds)-1 == i || objbest.MatchID != objLiveOdds[i+1].MatchID || objbest.CategoryID != objLiveOdds[i+1].CategoryID {
			geniusMarket.MatchID = objbest.MatchID
			geniusMarket.CategoryName = objbest.CategoryName
			geniusMarketOdds = append(geniusMarketOdds, geniusMarket)
			geniusMarket = IntMarketInfo{}
			objOdd = OddsInfo{}
		}
	}

	return geniusMarketOdds
}

//MakingGeniusLiveMarketOddsSort :
func MakingGeniusLiveMarketOddsSort(objLiveOdds []GeniusOddsMarket, sportID int, typeVal string) []IntMarketInfo {

	var geniusMarketOdds []IntMarketInfo
	var geniusMarket IntMarketInfo

	test := map[string][]OddsInfo{}
	var testArr []string

	for i, objbest := range objLiveOdds {

		var objOdd OddsInfo

		// for other markets
		if objbest.MarketPrice != nil {

			objOdd.Name = objbest.ProviderInfo.Name
			objOdd.Icon = objbest.ProviderInfo.Icon
			objOdd.ProviderSequence = objbest.ProviderInfo.GeniusOddsSequence

			objOdd.MarketID = objbest.ProviderMarketID
			objOdd.OpenOdds = objbest.MarketFlucPrice
			objOdd.NewOdds = objbest.MarketPrice
			objOdd.MarketInternalID = int(objbest.MarketID)
			objOdd.DisplayName = objbest.DisplayName
			objOdd.CategoryName = objbest.CategoryName
			objOdd.Flucs = objbest.Flucs
			objOdd.MarketName = objbest.MarketName
			objOdd.TypeVal = "market"

			//condition for line
			if strings.Contains(objbest.CategoryName, "Line") && objbest.MarketVal != nil {
				objOdd.OpenLine = objbest.MarketFlucVal
				objOdd.NewLine = objbest.MarketVal
			} else if strings.Contains(objbest.CategoryName, "Total") && objbest.MarketVal != nil {
				objOdd.OpenTotal = objbest.MarketFlucVal
				objOdd.NewTotal = objbest.MarketVal
			}

			if _, ok := test[objbest.DisplayName]; !ok {
				testArr = append(testArr, objbest.DisplayName)
			}
			test[objbest.DisplayName] = append(test[objbest.DisplayName], objOdd)

		}

		if len(objLiveOdds)-1 == i || objbest.CategoryName != objLiveOdds[i+1].CategoryName {

			for _, value := range testArr {
				if val, ok := test[value]; ok {
					sort.Sort(MatchMarketSort(val))
					geniusMarket.AnyMarketOdds = append(geniusMarket.AnyMarketOdds, val...)
				}
			}

			geniusMarket.CategoryName = objbest.CategoryName
			geniusMarket.GroupName = objbest.GroupName
			geniusMarketOdds = append(geniusMarketOdds, geniusMarket)

			geniusMarket = IntMarketInfo{}

			test = map[string][]OddsInfo{}
			testArr = []string{}
		}

	}

	return geniusMarketOdds
}

//GetBestMatchID :
func GetBestMatchID(bestMatches []OddsInfo, sportID, leagueID int, typeVal string) int {
	var matchID int
	for _, match := range bestMatches {
		if match.SportID == sportID && match.LeagueID == leagueID {
			matchID = match.MatchID
		}
	}
	return matchID
}

//GetPlungeMatchID :
func GetPlungeMatchID(bestMatches []GeniusOddsPlunge, sportID, leagueID int, typeVal string) (map[int64][]GeniusOddsPlunge, []string) {
	objPlugeMatches := map[int64][]GeniusOddsPlunge{}
	var matchIDs []string
	for _, match := range bestMatches {
		if match.SportID == sportID && match.LeagueID == leagueID {
			objPlugeMatches[match.MatchID] = append(objPlugeMatches[match.MatchID], match)
			matchIDs = append(matchIDs, strconv.Itoa(int(match.MatchID)))

		}
	}
	return objPlugeMatches, matchIDs
}

// MakingLiveOddsChangeSort :
func MakingLiveOddsChangeSort(liveOdds []FixtureOdds, sportID int) map[int64][]GeniusOddsPlunge {

	liveOddsChange := map[int64][]GeniusOddsPlunge{}
	var liveBothOddsChange []GeniusOddsPlunge

	for i, odds := range liveOdds {

		var liveHomeOddChange GeniusOddsPlunge
		var liveAwayOddChange GeniusOddsPlunge

		liveHomeOddChange.Plunge.OpenOdds = odds.HomePlunge.OpenOdds
		liveHomeOddChange.Plunge.NewOdds = odds.HomePlunge.NewOdds
		liveHomeOddChange.Plunge.ChangePercentage = odds.HomePlunge.ChangePercentage
		liveHomeOddChange.ProviderInfo = odds.ProviderInfo
		liveHomeOddChange.Plunge.Flucs = odds.HomePlunge.Flucs
		liveHomeOddChange.Plunge.OddsType = "home"

		liveAwayOddChange.Plunge.OpenOdds = odds.AwayPlunge.OpenOdds
		liveAwayOddChange.Plunge.NewOdds = odds.AwayPlunge.NewOdds
		liveAwayOddChange.Plunge.ChangePercentage = odds.AwayPlunge.ChangePercentage
		liveAwayOddChange.ProviderInfo = odds.ProviderInfo
		liveAwayOddChange.Plunge.Flucs = odds.AwayPlunge.Flucs
		liveAwayOddChange.Plunge.OddsType = "away"

		if *odds.HomePlunge.ChangePercentage < 0 && math.Abs(*odds.HomePlunge.ChangePercentage) > 15 {
			liveBothOddsChange = append(liveBothOddsChange, liveHomeOddChange)
		}
		if *odds.AwayPlunge.ChangePercentage < 0 && math.Abs(*odds.AwayPlunge.ChangePercentage) > 15 {
			liveBothOddsChange = append(liveBothOddsChange, liveAwayOddChange)
		}

		if ((len(liveOdds)-1) == i || odds.MatchID != liveOdds[i+1].MatchID) && len(liveBothOddsChange) > 0 {

			// Sort the match wise best plunge
			sort.Sort(PlungeSort(liveBothOddsChange))

			var livefinalOddsChange []GeniusOddsPlunge
			for _, val := range liveBothOddsChange {
				if liveBothOddsChange[0].Plunge.OddsType == val.Plunge.OddsType {
					livefinalOddsChange = append(livefinalOddsChange, val)
				}
			}

			liveOddsChange[odds.MatchID] = livefinalOddsChange
			liveBothOddsChange = []GeniusOddsPlunge{}
		}
	}

	return liveOddsChange
}

// MakingGeniusLiveOddsSortBgProc :
func MakingGeniusLiveOddsSortBgProc(objLiveOdds []GeniusOddsMarket, sportID int, typeVal string) []IntMarketInfo {

	var geniusMarketOdds []IntMarketInfo
	var geniusMarket IntMarketInfo

	var objHomeMarketOdds, objAwayMarketOdds, objAnyMarketOdds []OddsInfo

	for i, objbest := range objLiveOdds {

		var objOdd OddsInfo

		// for other markets
		if objbest.MarketPrice != nil {

			if objbest.HomeTeamID == objbest.MarketTeamID {
				objOdd.NewOdds = objbest.MarketPrice
				objHomeMarketOdds = append(objHomeMarketOdds, objOdd)

			} else if objbest.AwayTeamID == objbest.MarketTeamID {
				objOdd.NewOdds = objbest.MarketPrice
				objAwayMarketOdds = append(objAwayMarketOdds, objOdd)

			} else if objbest.MarketTeamID == 0 {
				objOdd.NewOdds = objbest.MarketPrice
				objAnyMarketOdds = append(objAnyMarketOdds, objOdd)
			}
		}

		if len(objLiveOdds)-1 == i || objbest.MatchID != objLiveOdds[i+1].MatchID {

			sort.Sort(MatchMarketSort(objHomeMarketOdds))
			sort.Sort(MatchMarketSort(objAwayMarketOdds))
			sort.Sort(MatchMarketSort(objAnyMarketOdds))
			geniusMarket.HomeMarketOdds = append(geniusMarket.HomeMarketOdds, objHomeMarketOdds...)
			geniusMarket.AwayMarketOdds = append(geniusMarket.AwayMarketOdds, objAwayMarketOdds...)
			geniusMarket.AnyMarketOdds = append(geniusMarket.AnyMarketOdds, objAnyMarketOdds...)
			geniusMarket.MatchID = objbest.MatchID
			geniusMarketOdds = append(geniusMarketOdds, geniusMarket)
			objHomeMarketOdds, objAwayMarketOdds, objAnyMarketOdds = []OddsInfo{}, []OddsInfo{}, []OddsInfo{}
			geniusMarket = IntMarketInfo{}
			objOdd = OddsInfo{}
		}
	}

	return geniusMarketOdds
}
