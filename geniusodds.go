package isg

// IntMarketInfo :
type IntMarketInfo struct {
	MatchID        int64
	CategoryName   string
	GroupName      string
	HomeMarketOdds []OddsInfo
	AwayMarketOdds []OddsInfo
	AnyMarketOdds  []OddsInfo
}

// GeniusOddsSportMatch :
type GeniusOddsSportMatch struct {
	Sport []GeniusOddsMatch `json:"sport,omitempty"`
}

// GeniusOddsMatch :
type GeniusOddsMatch struct {
	SportID   string                  `json:"sport_id,omitempty"`
	SportName string                  `json:"sport_name,omitempty"`
	SportURL  string                  `json:"sport_url,omitempty"`
	Leagues   []GeniusOddsMatchLeague `json:"leagues,omitempty"`
}

// GeniusOddsMatchLeague :
type GeniusOddsMatchLeague struct {
	Leaguename string                  `json:"league_name,omitempty"`
	LeagueURL  string                  `json:"league_url,omitempty"`
	Matches    []GeniusOddsMatchesInfo `json:"matches,omitempty"`
}

// GeniusOddsMatchesInfo :
type GeniusOddsMatchesInfo struct {
	MatchID         int64              `json:"match_id,omitempty"`
	LocalDate       string             `json:"local_date,omitempty"`
	LocalTime       string             `json:"local_time,omitempty"`
	MatchDate       string             `json:"match_date"`
	MatchTime       string             `json:"match_time"`
	TimeZone        string             `json:"timezone"`
	Weather         string             `json:"weather"`
	DayNight        string             `json:"playing,omitempty"`
	Status          string             `json:"match_status,omitempty"`
	IsReschedule    int                `json:"is_reschedule"`
	IsPlunge        string             `json:"plunge_team,omitempty"`
	Round           SportRound         `json:"round_week,omitempty"`
	HomeTeamInfo    Team               `json:"home,omitempty"`
	AwayTeamInfo    Team               `json:"away,omitempty"`
	VenueInfo       Venue              `json:"venue,omitempty"`
	Groups          []GeniusGroups     `json:"groups,omitempty"`
	Market          []GeniusMarketOdds `json:"markets,omitempty"`
	GeniusOddsPlung *MarketOddsList    `json:"plunge_odds,omitempty"`
}

// GeniusGroups :
type GeniusGroups struct {
	GroupName string             `json:"group_name,omitempty"`
	Market    []GeniusMarketOdds `json:"markets,omitempty"`
}

// GeniusMarketOdds :
type GeniusMarketOdds struct {
	MarketName        string                `json:"market_name,omitempty"`
	MarketOption      []GeniusMarketOptions `json:"market_options,omitempty"`
	MatchMarketOption []MarketOddsList      `json:"options,omitempty"`
}

// GeniusMarketOptions :
type GeniusMarketOptions struct {
	MarketType           string          `json:"market_type,omitempty"`
	GeniusHomeMarketOdds *MarketOddsList `json:"home,omitempty"`
	GeniusAwayMarketOdds *MarketOddsList `json:"away,omitempty"`
	GeniusAnyMarketOdds  *MarketOddsList `json:"others,omitempty"`
}

// MarketOddsList :
type MarketOddsList struct {
	MarketType   string     `json:"display_name,omitempty"`
	AbbrName     string     `json:"abbr_name,omitempty"`
	TeamInfo     *Team      `json:"team,omitempty"`
	ProviderList []OddsInfo `json:"providers,omitempty"`
}

// ExtProvider :
type ExtProvider struct {
	Name        *string   `json:"name,omitempty"`
	URL         *string   `json:"url,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Fluctuation []float64 `json:"fluc,omitempty"`
}
