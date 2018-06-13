/*
Package isg - iSport Genius Data Library
*/
package isg

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/leekchan/timeutil"
)

const (
	ISGErrGeneralPrefix  string = "ISGError: "
	ISGErrBadInputPrefix string = "ISGBadInputError: "
)

type LinkResource struct {
	Name     string `json:"device,omitempty"`
	PageName string `json:"name,omitempty"`
	Resource string `json:"url"`
}

type LinkResourceArray struct {
	Name     string         `json:"page,omitempty"`
	Resource []LinkResource `json:"resources"`
}

// FunfactsMatches :
type LinkResourceFunfacts struct {
	FunfactList []FunFact `json:"funfacts,omitempty"`
}

// LinkResourceAnalysis :
type LinkResourceAnalysis struct {
	Analysis string `json:"match_preview,omitempty"`
}

type Provider struct {
	ProviderId         string
	Name               string
	URL                string
	Icon               string
	IsTDProvider       bool
	ForPredictor       bool
	GeniusOddsSequence int
}

type Customer struct {
	Id   int
	Name string
	UUID string
}

type ProductInfo struct {
	ID          string
	ProductName string
	UUID        string
	ProviderID  string
	Icon        string
	ProductURL  string
	CustomerID  string
}

// RequestJSON : json will be passed by sportsbet for publish the event
type RequestJSON struct {
	EventID       int    `json:"eventid"`
	Sports        string `json:"sport"`
	League        string `json:"league"`
	HomeTeam      string `json:"team1"`
	AwayTeam      string `json:"team2"`
	EventDateTime string `json:"eventdatetime"`
}

//Sport :
type Sport struct {
	SportID          string `json:"sport_id"` // Unique two-character code for the sport e.g. "RL"
	SportName        string `json:"sport_name"`
	SportInternalID  int    `json:"sport_internal_id,omitempty"`
	TableNameSeasons string `json:"-"`
	TableNameMatches string `json:"-"`
	TableNamePlayers string `json:"-"`
	SportAPICode     string `json:"-"`
	SportURL         string `json:"sport_url,omitempty"`
	SportLogo        string `json:"sport_logo"`
}

//League :
type League struct {
	LeagueInternalID int     `json:"league_internal_id,omitempty"` // Internal league_id e.g. 1
	LeagueID         string  `json:"-"`                            // Official API ID e.g. "01"
	LeagueEntityKey  string  `json:"-"`                            // Unique League ID for identification e.g. "afl", "nrl", "8rm9ay"
	LeagueName       string  `json:"league_name"`
	LeagueURL        string  `json:"league_url,omitempty"`
	Seasons          *Season `json:"seasons,omitempty"`
}

//Season :
type Season struct {
	SportID          string `json:"-"`
	SportInternalID  int    `json:"-"`
	LeagueID         string `json:"-"`
	LeagueInternalID int    `json:"-"`
	SeasonID         string `json:"-"`
	SeasonInternalID int    `json:"season_id,omitempty"`
	SeasonName       string `json:"season"`
	SeasonURL        string `json:"season_url,omitempty"`
}

//Team :
type Team struct {
	TeamID                 string `json:"id"` // Official API ID e.g. kduw36b4
	TeamInternalID         string `json:"-"`  // Internal DB ID e.g. 1428
	TeamName               string `json:"name,omitempty"`
	FullName               string `json:"full_name,omitempty"`
	TeamShortName          string `json:"short_teamname,omitempty"`
	TeamAltName1           string `json:"-"`
	TeamAltName2           string `json:"-"`
	Abbreviation           string `json:"abbr,omitempty"`
	Conference             string `json:"conference,omitempty"`
	Division               string `json:"division,omitempty"`
	Ranking                string `json:"rank,omitempty"`
	DivisionRanking        string `json:"division_rank,omitempty"`
	DivisionDisplayRanking string `json:"division_display_rank,omitempty"`
	TeamURL                string `json:"team_url,omitempty"`
	TeamSBURL              string `json:"sb_team_url,omitempty"`
	TeamFlag               string `json:"icon,omitempty"`
	SortVal                int    `json:"-"`
	TeamColor              string `json:"color,omitempty"`
	TeamGroup              string `json:"group,omitempty"`
	Pitcher                string `json:"pitcher,omitempty"`
	Points                 int    `json:"-"`
	TeamReverseName        string `json:"-"`
	Country                string `json:"country,omitempty"`
	CountryID              int    `json:"-"`
	TeamNickName           string `json:"nick_name,omitempty"`
	GroupRanking           string `json:"group_rank,omitempty"`
}

type SportRound struct {
	RoundID        int    `json:"round_id,omitempty"`
	Name           string `json:"round_name,omitempty"`
	RoundCode      string `json:"round"`
	URL            string `json:"round_url,omitempty"`
	ShortRoundName string `json:"short_round_name,omitempty"`
}

type SportWeek struct {
	WeekID   int    `json:"week_id,omitempty"`
	Week     string `json:"week,omitempty"`
	WeekName string `json:"week_name,omitempty"`
	WeekURL  string `json:"week_url,omitempty"`
}

type SportSeasonType struct {
	SeasontypeID   int    `json:"season_type_id,omitempty"`
	SeasonType     string `json:"season_type,omitempty"`
	SeasonTypeName string `json:"season_type_name,omitempty"`
	SeasonTypeUrl  string `json:"season_type_url,omitempty"`
}

//NRLRounds
type NRLRounds struct {
	SeasonID int
	MaxRound int
}

//Market :
type Market struct {
	MarketID                 string
	MarketShortName          string
	MarketName               string
	MarketInternalID         int
	MarketCategoryInternalID int
	MarketCategoryName       string
}

//AllMatchDetail :
type AllMatchDetail struct {
	RoundID  int
	MaxScore int
	Count    int
	Season   int
}

// PortalEntity : It defines the common structure of O/U results teamwise
type PortalEntity struct {
	SportName   string
	EntityID    int
	EntityAPIID string
	EntityName  string
	SortVal     int
	URL         string
}

// PointsAdjustments : Adjust points for ranking according to sports
type PointsAdjustments struct {
	SportID  int
	SeasonID int
	TeamID   string
	Points   int
}

// ErrorAlertGroup :
type ErrorAlertGroup struct {
	AlertID   int
	GroupID   int
	AlertName string
	GroupName string
	UserID    string
}

// ErrorLogRecord :
type ErrorLogRecord struct {
	MatchID        int
	SeasonID       int
	LeagueID       int
	RoundWeek      string
	HomeTeam       string
	AwayTeam       string
	MatchDate      string
	MatchTime      string
	ScriptName     string
	ErroMsg        string
	Provider       string
	ServerLocation string
	PreRun         string
	LastRun        string
	Frequency      string
	Dateadded      string
	ProviderName   string
	AlertName      string
}

// Ycount :  count, roundid and seasonid struct for rugby-union
type Ycount struct {
	SeasonID int
	RoundID  int
	Count    int
}

//-------------------------------------------------------------------------
// Common Structs to be moved in Common.go
//-------------------------------------------------------------------------

//MatchInfo : holds all information about a match and common for all sports
type MatchInfo struct {
	MatchID                int
	SeasonID               int
	Season                 string
	RoundURL               string
	HomeTeamID             string
	HomeTeamInternalID     int
	HomeTeamName           string
	HomeTeamFilterName     string
	HomeTeamAbbr           string
	HomeTeamIcon           string
	HomeTeamColor          string
	HomeTeamURL            string
	HomeTeamSBURL          string
	AwayTeamID             string
	AwayTeamInternalID     int
	AwayTeamName           string
	AwayTeamFilterName     string
	AwayTeamAbbr           string
	AwayTeamIcon           string
	AwayTeamColor          string
	AwayTeamURL            string
	AwayTeamSBURL          string
	LocalDate              string
	LocalTime              string
	MatchDate              string
	MatchTime              string
	VenueID                string
	VenueFilterName        string
	VenueInternalID        int
	MatchWeather           sql.NullString
	MatchDayNight          sql.NullString
	MatchStatus            string
	MatchPreview           sql.NullString
	HomeTeamRank           string
	AwayTeamRank           string
	TournamentURL          string
	TournamentFilterName   string
	TournamentCountryName  string
	Status                 string
	TournamentID           string
	TournamentCountryID    string
	SurfaceName            string
	SurfaceURL             string
	HomeTeamFavUnd         string
	AwayTeamFavUnd         string
	Player1HandPostion     sql.NullString
	Player2HandPostion     sql.NullString
	HomeTeamDayRest        sql.NullInt64
	AwayTeamDayRest        sql.NullInt64
	HomeTeamState          sql.NullString
	AwayTeamState          sql.NullString
	VenueState             sql.NullString
	CurrentSeasonID        sql.NullInt64
	RoundID                sql.NullInt64
	HomeCountryInternalID  sql.NullInt64
	AwayCountryInternalID  sql.NullInt64
	HomeCountry            sql.NullString
	AwayCountry            sql.NullString
	WeekDay                string
	VenueCountry           sql.NullString
	VenueCountryInternalID sql.NullInt64
	OptionalMatchInfo      interface{} //other match info for particular sports
	HomeContinent          sql.NullString
	AwayContinent          sql.NullString
	VenueContinent         sql.NullString
	OptionalMatchInfo2     interface{} //other match info for particular sports
	HomeHMSTINTST          sql.NullString
	AwayHMSTINTST          sql.NullString
	HomeConference         sql.NullString
	AwayConference         sql.NullString
	HomeDivision           sql.NullString
	AwayDivision           sql.NullString
	HomeTeamShortName      sql.NullString
	AwayTeamShortName      sql.NullString
	HomePastH2HResult      sql.NullString
	AwayPastH2HResult      sql.NullString
	HomeTeamNickName       sql.NullString
	AwayTeamNickName       sql.NullString
}

// RoundWeek : holds all information for a round or week
type RoundWeek struct {
	RoundWeekID   int
	RoundWeekName string
	RoundWeekURL  string
}

// Venue : holds all information for a venue
type Venue struct {
	VenueID         string `json:"id"`
	VenueInternalID int    `json:"-"`
	VenueName       string `json:"-"`
	VenueFilterName string `json:"name"`
	VenueURL        string `json:"venue_url,omitempty"`
	VenueCountry    string `json:"country,omitempty"`
	VenueCity       string `json:"city,omitempty"`
	VenueState      string `json:"state,omitempty"`
	Latitude        string `json:"latitude,omitempty"`
	Longitude       string `json:"longitude,omitempty"`
}

// Tips : holds all information for a tips by macth
type Tips struct {
	TipsTitle string `json:"title,omitempty"`
	Option1   string `json:"option1_title,omitempty"`
	Option2   string `json:"option2_title,omitempty"`
	Tips      string `json:"tip,omitempty"`
	Roughie   string `json:"roughie,omitempty"`
}

// Odds :
type Odds struct {
	HomeH2HMarket       string                `json:"home_h2h_market,omitempty"`
	AwayH2HMarket       string                `json:"away_h2h_market,omitempty"`
	DrawH2HMarket       string                `json:"draw_h2h_market,omitempty"`
	HomeOdds            *float64              `json:"home_odds"`
	AwayOdds            *float64              `json:"away_odds"`
	DrawOdds            *float64              `json:"draw_odds,omitempty"`
	HomeLineMarket      string                `json:"home_line_market,omitempty"`
	AwayLineMarket      string                `json:"away_line_market,omitempty"`
	HomeLine            *float64              `json:"home_line,omitempty"`
	HomeLineOdds        *float64              `json:"home_line_odds,omitempty"`
	AwayLine            *float64              `json:"away_line,omitempty"`
	AwayLineOdds        *float64              `json:"away_line_odds,omitempty"`
	OverMarket          string                `json:"over_market,omitempty"`
	UnderMarket         string                `json:"under_market,omitempty"`
	ClosingTotal        *float64              `json:"closing_total,omitempty"`
	OverOdds            *float64              `json:"over_odds,omitempty"`
	UnderOdds           *float64              `json:"under_odds,omitempty"`
	HomeFirstOdds       *float64              `json:"-"`
	AwayFirstOdds       *float64              `json:"-"`
	DrawFirstOdds       *float64              `json:"-"`
	HomeOddsFluctuation *float64              `json:"home_odds_fluctuation,omitempty"`
	AwayOddsFluctuation *float64              `json:"away_odds_fluctuation,omitempty"`
	DrawOddsFluctuation *float64              `json:"draw_odds_fluctuation,omitempty"`
	OddsType            *string               `json:"-"`
	BttsYesMarket       *string               `json:"btts_yes_market,omitempty"`
	BttsNoMarket        *string               `json:"btts_no_market,omitempty"`
	BttsOddsYes         *float64              `json:"btts_odds_yes,omitempty"`
	BttsOddsNo          *float64              `json:"btts_odds_no,omitempty"`
	HomeBigWinLittleWin BigWinLittleWinMarket `json:"-"`
	AwayBigWinLittleWin BigWinLittleWinMarket `json:"-"`
	HomeWireToWire      WireToWireMarket      `json:"-"`
	AwayWireToWire      WireToWireMarket      `json:"-"`
	AnyWireToWire       WireToWireMarket      `json:"-"`
	HomeHTFT            HTFTMarket            `json:"-"`
	AwayHTFT            HTFTMarket            `json:"-"`
	AnyHTFTOdds         *float64              `json:"-"`
	AnyHTFTMarketID     *string               `json:"-"`
	HomeMinWin          MinWinMarket          `json:"-"`
	AwayMinWin          MinWinMarket          `json:"-"`
	MinWinTieOdds       *float64              `json:"-"`
	MinWinTieMarketID   *string               `json:"-"`
	HomeTriBet          TriBetMarket          `json:"-"`
	AwayTriBet          TriBetMarket          `json:"-"`
	AnyTriBet           TriBetMarket          `json:"-"`
}

// FixtureOdds :
type FixtureOdds struct {
	MatchID                int64             `json:"-"`
	HomeOdds               *float64          `json:"home_odds,omitempty"`
	AwayOdds               *float64          `json:"away_odds,omitempty"`
	DrawOdds               *float64          `json:"draw_odds,omitempty"`
	HomeLine               *float64          `json:"home_line,omitempty"`
	HomeLineOdds           *float64          `json:"home_line_odds,omitempty"`
	AwayLine               *float64          `json:"away_line,omitempty"`
	AwayLineOdds           *float64          `json:"away_line_odds,omitempty"`
	ClosingTotal           *float64          `json:"closing_total,omitempty"`
	OverOdds               *float64          `json:"over_odds,omitempty"`
	UnderOdds              *float64          `json:"under_odds,omitempty"`
	BttsYesOdds            *float64          `json:"btts_yes_odds,omitempty"`
	BttsNoOdds             *float64          `json:"btts_no_odds,omitempty"`
	FirstHomeOdds          *float64          `json:"first_home_odds,omitempty"`
	FirstAwayOdds          *float64          `json:"first_away_odds,omitempty"`
	FirstDrawOdds          *float64          `json:"first_draw_odds,omitempty"`
	FirstHomeLine          *float64          `json:"first_home_line,omitempty"`
	FirstHomeLineOdds      *float64          `json:"first_home_line_odds,omitempty"`
	FirstAwayLine          *float64          `json:"first_away_line,omitempty"`
	FirstAwayLineOdds      *float64          `json:"first_away_line_odds,omitempty"`
	FirstClosingTotal      *float64          `json:"first_closing_total,omitempty"`
	FirstOverOdds          *float64          `json:"first_over_odds,omitempty"`
	FirstUnderOdds         *float64          `json:"first_under_odds,omitempty"`
	FirstBttsYesOdds       *float64          `json:"first_btts_yes_odds,omitempty"`
	FirstBttsNoOdds        *float64          `json:"first_btts_no_odds,omitempty"`
	HomeBigWinOdds         *float64          `json:"home_big_win_odds,omitempty"`
	FirstHomeBigWinOdds    *float64          `json:"first_home_big_win_odds,omitempty"`
	HomeLittleWinOdds      *float64          `json:"home_little_win_odds,omitempty"`
	FirstHomeLittleWinOdds *float64          `json:"first_home_little_win_odds,omitempty"`
	AwayBigWinOdds         *float64          `json:"away_big_win_odds,omitempty"`
	FirstAwayBigWinOdds    *float64          `json:"first_away_big_win_odds,omitempty"`
	AwayLittleWinOdds      *float64          `json:"away_little_win_odds,omitempty"`
	FirstAwayLittleWinOdds *float64          `json:"first_away_little_win_odds,omitempty"`
	ProviderInfo           Provider          `json:"-"`
	HomePlunge             FixturePlungeOdds `json:"-"`
	AwayPlunge             FixturePlungeOdds `json:"-"`
	DrawPlunge             FixturePlungeOdds `json:"-"`
	HomeFlucs              []*float64        `json:"-"`
	AwayFlucs              []*float64        `json:"-"`
	HomeH2HMarket          string            `json:"home_h2h_market,omitempty"`
	AwayH2HMarket          string            `json:"away_h2h_market,omitempty"`
	DrawH2HMarket          string            `json:"draw_h2h_market,omitempty"`
	HomeH2HOption          string            `json:"home_h2h_option,omitempty"`
	AwayH2HOption          string            `json:"away_h2h_option,omitempty"`
	DrawH2HOption          string            `json:"draw_h2h_option,omitempty"`
	MarketName             string            `json:"market_type,omitempty"`
}

// GeniusOddsPlunge :
type GeniusOddsPlunge struct {
	MatchID      int64
	SportID      int
	LeagueID     int
	Plunge       FixturePlungeOdds
	ProviderInfo Provider
}

// TeamLists :
type TeamLists struct {
	Sports   Sport  `json:"sport"`
	League   League `json:"league"`
	TeamList []Team `json:"teams"`
}

// SportsCurrentRound :
type SportsCurrentRound struct {
	Sports Sport  `json:"sport"`
	League League `json:"league"`
	Season string `json:"season"`
	Round  string `json:"current_round_week"`
}

// LevelTournament :
type LevelTournament struct {
	Sports     Sport        `json:"sport"`
	League     League       `json:"league"`
	Tournament []Tournament `json:"tournaments"`
}

// SportsFormStreak :
type SportsFormStreak struct {
	Form       string
	Streak     string
	MatchCount int
}

// DateString returns today's date in string format.
func DateString() string {
	datevalue := time.Now()
	dateStr := timeutil.Strftime(&datevalue, "%Y-%m-%d")
	return dateStr
}

// NowString returns today's datetime in YYYY-MM-DD hh:mm:ss format.
func NowString() string {
	datevalue := time.Now()
	dateStr := timeutil.Strftime(&datevalue, "%Y-%m-%d %H:%M:%S")
	return dateStr
}

// Round rounds the number to the desired decimal place.
// e.g. Round(-3.745, 2) == -3.75
//      Round(-3.745, 1) == -3.7
// Rules according to http://www.aaamath.com/est-dec-round.htm
// Source: https://gist.github.com/DavidVaini/10308388
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64

	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)

	if val < 0.0 {
		div = div * -1.0

		if div >= roundOn {
			round = math.Floor(digit)
		} else {
			round = math.Ceil(digit)
		}
	} else {
		if div >= roundOn {
			round = math.Ceil(digit)
		} else {
			round = math.Floor(digit)
		}
	}

	newVal = round / pow
	return
}

// StringCombos all combinations of a string array's items (up to r iterations)
// and put them together in an array with separator used on each combo string.
func StringCombos(iterable []string, r int, separator string) []string {
	// Borrowed from https://play.golang.org/p/JEgfXR2zSH
	results := []string{}
	pool := iterable
	n := len(pool)

	if r > n {
		return nil
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	stringList := make([]string, r)
	for i, el := range indices {
		stringList[i] = pool[el]
	}

	//fmt.Println(strings.Join(stringList[:], separator))
	results = append(results, strings.Join(stringList[:], separator))

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i -= 1 {
		}

		if i < 0 {
			return results
		}

		indices[i] += 1
		for j := i + 1; j < r; j += 1 {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i += 1 {
			stringList[i] = pool[indices[i]]
		}
		//fmt.Println(strings.Join(stringList[:], separator))
		results = append(results, strings.Join(stringList[:], separator))

	}

	return results

}

// Compress returns a compressed (gzip) array of bytes.
func Compress(data []byte) []byte {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, flate.BestCompression)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	_, err = w.Write([]byte(data))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	w.Close()
	return b.Bytes()
}

// CompressByLevel adds a degree of compression level for flexibility.
func CompressByLevel(data []byte, compression int) []byte {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, compression)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	_, err = w.Write([]byte(data))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	w.Close()
	return b.Bytes()
}

// Extract returns a decompressed array of (gzipped) bytes.
func Extract(data []byte) []byte {
	var b bytes.Buffer
	b.Write(data)
	r, err := gzip.NewReader(&b)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	results, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	r.Close()
	return results
}

func Reverse(strform []string) []string {
	for i := 0; i < len(strform)/2; i++ {
		j := len(strform) - i - 1
		strform[i], strform[j] = strform[j], strform[i]
	}
	return strform
}

func Even(number int) bool {
	return number%2 == 0
}

func SliceContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

/*
// GetVenueID finds the actual venue ID for the specified venue name/API ID.
func GetVenueID(venue string) (int, error) {
	return 0, nil
}

// GetEntityInternalID returns the internal integer ID of an entity based on its name/API ID.
func GetEntityInternalID(entity string) (int, error) {
	return 0, nil
}*/

func PositionMatch(matchvalue string, value string) bool {

	matched, _ := regexp.MatchString(strings.ToLower(matchvalue), strings.ToLower(value))
	return matched
}

func RemoveDuplicates(elements []int) []int {
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}

func generateScorePatterns(from int, to int, limit int, score string) bool {
	//var _ismatched = false
	for i := from; i <= limit; i++ {
		for j := to; j <= limit; j++ {
			newscore := strconv.FormatInt(int64(i), 10) + "-" + strconv.FormatInt(int64(j), 10)
			if newscore == score {
				//_ismatched = true
				return true
			}
		}
	}
	return false
}

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

func MakeCurrentStreak(streak string) string {
	var cnt int
	var firstletter string
	value := reverseform(streak)
	result := strings.SplitAfter(value, "")
	for _, c := range result {
		firstletter = result[0]
		if c == firstletter {
			cnt++
		} else {
			break
		}
	}
	return firstletter + strconv.Itoa(cnt)
}

func reverseform(s string) string {
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

//Months :
type Months struct {
	Name  string `json:"month"`
	Dayes []Day  `json:"days"`
}

//Day :
type Day struct {
	Day        int         `json:"day"`
	OnThisDays []OnThisDay `json:"details"`
}

//OnThisDay :
type OnThisDay struct {
	Year      int    `json:"year"`
	SportName string `json:"sport"`
	Icon      string `json:"icon"`
	OnThisDay string `json:"fact"`
}

//OnThisDayData :
type OnThisDayData struct {
	Day       sql.NullInt64
	Month     sql.NullInt64
	Year      sql.NullInt64
	Date      sql.NullString
	Sport     sql.NullString
	OnThisDay sql.NullString
}

//EventInfo :
type EventInfo struct {
	Sport                 Sport    `json:"sport"`
	League                League   `json:"league,omitempty"`
	MatchID               int      `json:"match_id,omitempty"`
	Season                []Season `json:"seasons,omitempty"`
	CurrentSeasonMatch    int      `json:"current_season_played"`
	RoundWeek             int      `json:"round_week"`
	TeamTotalPlayed       int      `json:"team_played"`
	ProviderID            int      `json:"provider_id,omitempty"`
	ProviderName          string   `json:"provider_name,omitempty"`
	HomeTeam              Team     `json:"home_team,omitempty"`
	AwayTeam              Team     `json:"away_team,omitempty"`
	EventRoundWeekURL     string   `json:"round_week_url,omitempty"`
	TournamentURL         string   `json:"tournament_url,omitempty"`
	TournamentFilterName  string   `json:"tournament_filter_name,omitempty"`
	TournamentCountryName string   `json:"tournament_country_name,omitempty"`
	MatchPlayIN           string   `json:"play_in,omitempty"`
	MatchDate             string   `json:"match_date,omitempty"`
	MatchStatus           string   `json:"match_status,omitempty"`
}

// EventInfoSbCache :
type EventInfoSbCache struct {
	StatusCode int       `json:"statusCode"`
	Event      EventInfo `json:"content,omitempty"`
	Message    string    `json:"message,omitempty"`
}

// BasketballFixtureSbCache :
type BasketballFixtureSbCache struct {
	StatusCode int                    `json:"statusCode"`
	Content    BasketballMatchFixture `json:"content"`
}

//EventMatchRecordInfo :
type EventMatchRecordInfo struct {
	Sport           Sport
	League          string
	LeagueURL       string
	LeagueID        int
	CurrentSeason   string
	CurrentSeasonID int
	LadderSeasonID  int
	MatchID         int
	Seasons         []Season
	ProviderID      int
}

//LOCAEST :
func LOCAEST() *time.Location {
	LocAEST, _ := time.LoadLocation("Australia/Melbourne")
	return LocAEST
}

var NumericLowNames = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}

var NumericTensNames = []string{"twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}

var NumericBigNames = []string{"thousand", "million", "billion"}

//convert999 : convert less than 1000 and greater than 99 into words
func convert999(num int) string {
	s1 := NumericLowNames[num/100] + " hundred"

	s2 := convert99(num % 100)
	if num <= 99 {
		return s2
	}
	if num%100 == 0 {
		return s1
	}
	return s1 + " " + s2
}

//convert99 : convert less than 99  into words
func convert99(num int) string {
	if num < 20 {
		return NumericLowNames[num]
	}

	s := NumericTensNames[num/10-2]
	if num == 0 {
		return s
	} else if num%10 == 0 {
		return s
	}
	return s + " " + NumericLowNames[num%10]
}

//ConvertNum2Words :
func ConvertNum2Words(num int) string {
	if num < 0 {
		return "negative " + ConvertNum2Words(-num)
	}

	if num <= 999 {
		return convert999(num)
	}

	s := ""
	t := 0
	for num > 0 {
		if num%1000 != 0 {
			s2 := convert999(num % 1000)
			if t > 0 {
				s2 = s2 + " " + NumericBigNames[t-1]
			}
			if s == "" {
				s = s2
			} else {
				s = s2 + ", " + s
			}
		}
		num /= 1000
		t++
	}
	return s
}

//FactorialDividend :
func FactorialDividend(n int64) int64 {
	if n < 0 {
		return 1
	}
	if n == 0 {
		return 1
	}
	return n * FactorialDividend(n-1)
}

type SportForSpoting struct {
	Sort  string
	Order string
	Teams map[int64]int
}

//SliceContainsString :
func SliceContainsString(s []string, e string) bool {
	for _, a := range s {
		if strings.TrimSpace(a) == strings.TrimSpace(e) {
			return true
		}
	}
	return false
}

// Sportlist
type Sportlist struct {
	SportInternalID int                `json:"-"`
	SportID         string             `json:"sportid"`
	SportName       string             `json:"sportname"`
	SportLogo       string             `json:"sport_logo"`
	SportURL        string             `json:"sport_url"`
	Leagues         []League           `json:"leagues,omitempty"`
	LeagueRegions   []LeagueRegion     `json:"regions,omitempty"`
	TournametList   []LeagueTournament `json:"tournament_leagues,omitempty"`
}

// LeagueRegion :
type LeagueRegion struct {
	RegionID   int      `json:"-"`
	RegionName string   `json:"region_name"`
	Leagues    []League `json:"leagues"`
}

// LeagueTournament :
type LeagueTournament struct {
	Leagues    League       `json:"league,omitempty"`
	Tournament []Tournament `json:"tournaments,omitempty"`
}

type FixturePlungeOdds struct {
	TeamID           int
	OddsType         string
	OpenOdds         *float64
	NewOdds          *float64
	ChangePercentage *float64
	Provider         string
	TeamOpenOdds     *float64
	Flucs            []*float64
}

// GeniusSportsMatch : result for each matches
type GeniusSportsMatch struct {
	SportInfo          Sport
	LeagueInfo         League
	MatchID            sql.NullInt64
	SeasonID           sql.NullInt64
	Season             sql.NullString
	RoundID            sql.NullString
	Round              sql.NullString
	RoundFullName      sql.NullString
	RoundURL           sql.NullString
	VenueID            sql.NullString
	VenueInternalID    sql.NullInt64
	VenueName          sql.NullString
	VenueURL           sql.NullString
	VenueCountry       sql.NullString
	VenueCity          sql.NullString
	HomeTeamID         sql.NullString
	HomeTeamInternalID sql.NullInt64
	HomeTeamName       sql.NullString
	HomeTeamFullName   sql.NullString
	HomeTeamAbbr       sql.NullString
	HomeTeamIcon       sql.NullString
	HomeTeamColor      sql.NullString
	HomeTeamGroup      sql.NullString
	HomeTeamURL        sql.NullString
	HomeTeamRank       sql.NullInt64
	HomeTeamShortName  sql.NullString
	AwayTeamID         sql.NullString
	AwayTeamInternalID sql.NullInt64
	AwayTeamName       sql.NullString
	AwayTeamFullName   sql.NullString
	AwayTeamAbbr       sql.NullString
	AwayTeamIcon       sql.NullString
	AwayTeamColor      sql.NullString
	AwayTeamGroup      sql.NullString
	AwayTeamURL        sql.NullString
	AwayTeamShortName  sql.NullString
	AwayTeamRank       sql.NullInt64
	MatchDate          sql.NullString
	MatchTime          sql.NullString
	TimeZone           sql.NullString
	CounterDate        sql.NullString
	CounterTime        sql.NullString
	MatchWeather       sql.NullString
	MatchDayNight      sql.NullString
	MatchStatus        sql.NullString
	MatchReschedule    int
	MatchOdds          FixtureOdds
	IntMatchOdds       []IntMarketInfo
	PlungeOddsList     []GeniusOddsPlunge
	MatchTeamRank      sql.NullInt64
	TypeVal            string
}

// for Listing on Sports Sequence for team rank is same
var SportsSequence = map[int]int{
	1:  1,
	2:  6,
	3:  5,
	4:  3,
	5:  8,
	6:  4,
	7:  2,
	8:  9,
	9:  7,
	10: 10,
}

// AutoFFEmailContents
type AutoFFEmailContents struct {
	HomeTeam    string
	AwayTeam    string
	MatchDate   string
	MatchTime   string
	Market      string
	Score       int
	Description string
}

// BigWinLittleWinMarket :
type BigWinLittleWinMarket struct {
	BigWinOdds        *float64
	BigWinMarketID    *string
	LittleWinOdds     *float64
	LittleWinMarketID *string
}

// WireToWireMarket :
type WireToWireMarket struct {
	WireToWireOdds   *float64
	WireToWireMarket *string
}

// HTFTMarket :
type HTFTMarket struct {
	HTFTOdds             *float64
	HTFTMarketID         *string
	TeamHTFTOdds         *float64
	TeamHTFTMarketID     *string
	DrawHTFTOdds         *float64
	DrawHTFTMarketID     *string
	TeamDrawHTFTOdds     *float64
	TeamDrawHTFTMarketID *string
}

// MinWinMarket :
type MinWinMarket struct {
	MinWin1Odds        *float64
	MinWin2Odds        *float64
	MinWinMoreOdds     *float64
	MinWin1MarketID    *string
	MinWin2MarketID    *string
	MinWinMoreMarketID *string
}

// TrieBetMarket :
type TriBetMarket struct {
	TriBetOdds   *float64
	TriBetMarket *string
}

// MarketOdds :
type MarketOdds struct {
	MarketName   string          `json:"market_name,omitempty"`
	MarketOption []MarketOptions `json:"market_options,omitempty"`
}

// MarketOptions :
type MarketOptions struct {
	MarketType        string    `json:"market_type,omitempty"`
	HomeMarketOdds    *OddsInfo `json:"home,omitempty"`
	AwayMarketOdds    *OddsInfo `json:"away,omitempty"`
	DrawyMarketOdds   *OddsInfo `json:"draw,omitempty"`
	AnyMarketOdds     *OddsInfo `json:"any,omitempty"`
	ClosingMarketOdds *OddsInfo `json:"closing_total,omitempty"`
	OverMarketOdds    *OddsInfo `json:"over,omitempty"`
	UnderMarketOdds   *OddsInfo `json:"under,omitempty"`
	BttsYesMarketOdds *OddsInfo `json:"yes,omitempty"`
	BttsNoMarketOdds  *OddsInfo `json:"no,omitempty"`
}

// OddsInfo :
type OddsInfo struct {
	ProviderInfo       *ExtProvider `json:"provider,omitempty"`
	Name               string       `json:"name,omitempty"`
	Icon               string       `json:"icon,omitempty"`
	OpenOdds           *float64     `json:"open_price,omitempty"`
	NewOdds            *float64     `json:"price,omitempty"`
	OpenLine           *float64     `json:"open_line,omitempty"`
	NewLine            *float64     `json:"current_line,omitempty"`
	OpenTotal          *float64     `json:"open_closing_total,omitempty"`
	NewTotal           *float64     `json:"current_closing_total,omitempty"`
	FlucPer            *float64     `json:"fluc_per,omitempty"`
	UpDownArrow        string       `json:"up_down_arrow,omitempty"`
	ProviderOrder      int          `json:"order,omitempty"`
	MarketID           string       `json:"market,omitempty"`
	Flucs              []*float64   `json:"fluc,omitempty"`
	MarketInternalID   int          `json:"-"`
	MarketName         string       `json:"market_name,omitempty"`
	DisplayName        string       `json:"-"`
	CategoryInternalID int          `json:"-"`
	CategoryName       string       `json:"-"`
	OppoTeamName       string       `json:"team_name,omitempty"`
	PlayerName         string       `json:"player_name,omitempty"`
	MatchID            int          `json:"-"`
	SportID            int          `json:"-"`
	LeagueID           int          `json:"-"`
	TeamID             int          `json:"-"`
	ProviderSequence   int          `json:"-"`
	TypeVal            string       `json:"-"`
}

// GeniusOddsMarket :
type GeniusOddsMarket struct {
	MatchID          int64
	ProviderInfo     Provider
	HomeTeamID       int64
	AwayTeamID       int64
	MarketID         int64
	MarketName       string
	CategoryID       int64
	CategoryName     string
	MarketTeamID     int64
	MarketPrice      *float64
	MarketVal        *float64
	MarketFlucPrice  *float64
	MarketFlucVal    *float64
	ProviderMarketID string
	Flucs            []*float64
	ContentID        int64
	GroupName        string
	ParentID         int64
	DisplayName      string
	Sequence         int
	ISGapiID         string
}

//CheckValueInArray : check val in existing array
func CheckValueInArray(arr []string, val string) bool {
	for _, val1 := range arr {
		if val1 == val {
			return true
		}
	}
	return false
}

// GeniusOdds :  Genius Odds
type GeniusOdds struct {
	HomeOdds            string                `json:"homeodds"`
	AwayOdds            string                `json:"awayodds"`
	DrawOdds            string                `json:"drawodds"`
	MatchID             string                `json:"matchid"`
	CorrectScoreDetails []CorrectScoreDetails `json:"correctscoredetails"`
}

// CorrectScoreDetails : Correct Score details
type CorrectScoreDetails struct {
	CorrectScore       string `json:"correctscore"`
	CorrectScoreOdds   string `json:"correctscoreodds"`
	CorrectScoreMarket string `json:"correctscoremarket"`
}

//WeatherInfo :
type WeatherInfo struct {
	Weather     string `json:"weather,omitempty"`
	Temperature string `json:"temperature,omitempty"`
	WindSpeed   string `json:"wind_speed,omitempty"`
	Precip      string `json:"precip,omitempty"`
}

// GetDisplayRank :
func GetDisplayRank(str string) string {

	var outputStr string
	if str != "" {
		if strings.HasSuffix(str, "11") || strings.HasSuffix(str, "12") || strings.HasSuffix(str, "13") {
			outputStr = str + "th"
		} else if strings.HasSuffix(str, "1") {
			outputStr = str + "st"
		} else if strings.HasSuffix(str, "2") {
			outputStr = str + "nd"
		} else if strings.HasSuffix(str, "3") {
			outputStr = str + "rd"
		} else {
			outputStr = str + "th"
		}
	}
	return outputStr
}
