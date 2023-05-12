package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

// {"status":"SUCCESS","result":{"id":12345,"fullName":"John Smith","firstName":"John","lastName":"Smith",
// "shortAddress":"Boston, MA, US","gender":"MALE","age":32,
// "ratings":{"singles":"NR","singlesVerified":"NR","singlesProvisional":false,
// "doubles":"3.37","doublesVerified":"NR","doublesProvisional":false,"defaultRating":"DOUBLES"},
// "enablePrivacy":false,"isPlayer1":true,"verifiedEmail":true,"registered":true,
// "duprId":"3X7GM5","showRatingBanner":false,"sponsor":{}}}
type Ratings struct {
	Singles            string
	SinglesVerified    string
	SinglesProvisional bool
	Doubles            string
	DoublesVerified    string
	DoublesProvisional bool
	defaultRating      string
}

func (r *Ratings) Display() string {
	return r.Doubles
}

type APIPlayer struct {
	ID           uint
	DUPRId       string
	FullName     string
	FirstName    string
	LastName     string
	ShortAddress string
	gender       string
	age          int
	image_url    string
	email        string
	phone        string
	Ratings      Ratings
}

var dupr *DUPRClient

type DUPRClient struct {
	c           *resty.Client
	api_version string
	endpoint    string
	accessToken string
	debug       bool
}

// Helpful utility to build URL with known parts
func (dc *DUPRClient) build_url(url string) string {
	x := fmt.Sprintf("%s/%s", dc.endpoint, url)
	log.Println(x)
	return x
}

func (dc *DUPRClient) Get(url string) (*resty.Response, error) {
	return dc.c.R().Get(dc.build_url(url))
}

func (dc *DUPRClient) GetRequest() *resty.Request {
	r := dc.c.SetDebug(dc.debug).R()
	if dc.accessToken != "" {
		r = r.SetAuthToken(dc.accessToken)
	}
	return r
}

func NewDUPRClient() *DUPRClient {
	d := DUPRClient{
		api_version: "1.0",
		endpoint:    "https://api.dupr.gg",
	}
	//d.endpoint = "https://reqres.in"
	//d.api_version = "api"
	d.c = resty.New()
	return &d

}

func PrettyJson(data interface{}) string {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(val)
}

// Standard reply header, if Status is not "SUCCESS" there will
// be an error Message
type DUPRHeader struct {
	Status  string
	Message string
}

type LoginResultResponse struct {
	AccessToken  string
	RefreshToken string
}

type LoginResponse struct {
	DUPRHeader
	Result LoginResultResponse
}

type GetPlayerResponse struct {
	DUPRHeader
	Result APIPlayer
}

func handlePaging(r *PagedResult) int {
	// Handle results that are paged.
	// use like this:
	//
	// while offset is not None:
	//    dupr_get
	//    offset, hits = handle_paging(response.json())
	if r.Offset+r.Limit < r.Total {
		// there is more
		return (int)(r.Offset + r.Limit)
	} else {
		return -1
	}
}

type PagedResult struct {
	Offset uint
	Limit  uint
	Total  uint
	Hits   []interface{}
}

type PagedResponse struct {
	DUPRHeader
	Result PagedResult
}

func (dc *DUPRClient) Auth(username string, password string) (*resty.Response, error) {
	log.Printf("Logging in %s\n", username)

	data := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    username,
		Password: password,
	}

	log.Println(data)

	var result LoginResponse

	resp, err := dc.c.R().SetResult(&result).SetBody(&data).Post(dc.build_url("auth/v1.0/login/"))
	if err != nil {
		log.Println("error")
		return nil, nil
	}

	b, _ := json.Marshal(result)
	os.WriteFile("dupr.json", b, 0644)

	return resp, err
}

func (dc *DUPRClient) LoadTokens() error {
	// Load token for access without relogging in
	bytes, err := os.ReadFile("dupr.json")
	if err != nil {
		fmt.Println("Unable to load token file!")
		return err
	}

	var result LoginResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		fmt.Println("Unable to read token file data")
		return err
	}
	dc.accessToken = result.Result.AccessToken
	if dc.accessToken == "" {
		log.Println("Access token is empty")
	}
	return nil
}

func login(username string, password string, debug bool) {

	dupr = NewDUPRClient()
	dupr.debug = debug
	err := dupr.LoadTokens()
	if err == nil {
		log.Println("Access token loaded")
		return
	}
	_, err = dupr.Auth(username, password)
	if err != nil {
		log.Fatal(err)
	}

}

func GetPlayer(pid string) {
	// NOTE: This call does not require auth token..?!
	url := dupr.build_url("player/v1.0/" + pid)
	var result GetPlayerResponse
	_, err := dupr.c.SetDebug(dupr.debug).R().SetResult(&result).Get(url) // SetAuthToken(dupr.accessToken).Get(url)
	if err != nil {
		log.Fatal(err)
	}
}

type SortSpec struct {
	Order     string `json:"order"`
	Parameter string `json:"parameter"`
}

type MembersByClubPageRequest struct {
	// TODO: exclude with [""] crashed dupr, how to send exclude: [] ?
	// ANS: for now, exclude: null works
	//Exclude [1]string `json:"exclude"`
	//
	// This is different from the MemberMatchHistory because
	// this uses "filter" the other one uses "filters"
	Exclude []string `json:"exclude"`
	Limit   uint     `json:"limit"`
	Offset  uint     `json:"offset"`
	Query   string   `json:"query"`
	Filter  []string `json:"filter"`
	Sort    SortSpec `json:"sort"`
}

type MemberMatchHistoryPageRequest struct {
	Exclude []string `json:"exclude"`
	Limit   uint     `json:"limit"`
	Offset  uint     `json:"offset"`
	Query   string   `json:"query"`
	Filters []string `json:"filters"`
	Sort    SortSpec `json:"sort"`
}

/*
	{
	            "id": 4383648392,
	            "fullName": "John Smith",
	            "username": null,
	            "displayUsername": false,
	            "email": "jsmith@abcmail.com",
	            "verifiedEmail": true,
	            "isoAlpha2Code": "US",
	            "phone": "+16175551212",
	            "verifiedPhone": false,
	            "shortAddress": "Boston, MA, US",
	            "formattedAddress": "123 Pine Dr, Boston, MA 02123, USA",
	            "latitude": 41.4353212,
	            "longitude": -70.0022977,
	            "gender": "MALE",
	            "birthdate": "1991-01-31",
	            "age": 28,
	            "hand": null,
	            "imageUrl": null,
	            "singles": "NR",
	            "singlesVerified": "NR",
	            "singlesProvisional": false,
	            "doubles": "2.81",
	            "doublesVerified": "NR",
	            "doublesProvisional": false,
	            "defaultRating": "DOUBLES",
	            "distance": "1234.3 mi",
	            "distanceInMiles": 1234.3,
	            "enablePrivacy": false,
	            "status": "ACTIVE",
	            "created": null,
	            "clubId": 8310950964,
	            "roles": [
	               {
	                  "roleId": 2,
	                  "role": "PLAYER",
	                  "approvalStatus": "APPROVED",
	                  "clubId": 8311234567,
	                  "created": "2023-01-01T02:54:46.42406Z",
	                  "requestBy": 0,
	                  "joinType": null
	               }
	            ],
*/
type APIMember struct {
	// gorm.Model
	ID               uint64
	DUPRId           string
	FullName         string
	Username         string
	DisplayUsername  bool
	Email            string
	VerifiedEmail    bool
	Phone            string
	VerifiedPhone    bool
	ShortAddress     string
	FormattedAddress string
	Gender           string
	Birthday         string
	Age              int
	Image_url        string
	Ratings
}

type PagedMembersResponse struct {
	DUPRHeader
	Result PagedResult
	Hits   []APIMember
}

func GetMembersByClub(clubId string) ([]APIMember, error) {
	// /club/{club_id}/members/v1.0/all
	log.Printf("GetMemberByClub %s\n", clubId)
	var pageIn MembersByClubPageRequest
	// pageIn.Exclude[0] = ""
	pageIn.Query = "*"
	pageIn.Limit = 10
	pageIn.Sort.Order = "DESC"
	pageIn.Sort.Parameter = "fullNameSort"
	pageIn.Filter = nil

	log.Println(pageIn)
	dupr.LoadTokens()
	//goodJson := "{'exclude': [], 'limit': 20, 'offset': 0, 'query': '*'}"
	var result PagedResponse
	url := dupr.build_url("club/" + clubId + "/members/v1.0/all")

	var offset int = 0
	var members []APIMember

	for offset >= 0 {
		pageIn.Offset = uint(offset)
		_, err := dupr.GetRequest().SetBody(&pageIn).SetResult(&result).Post(url)
		if err != nil {
			log.Println("error")
			return nil, nil
		}
		offset = handlePaging(&result.Result)
		//log.Println(len(result.Result.Hits))
		//log.Println(result.Result.Hits[0])
		// TEMP
		if offset > 1000 {
			return members, nil
		}
		for _, h := range result.Result.Hits {
			var m APIMember
			d, _ := json.Marshal(h)
			json.Unmarshal(d, &m)
			members = append(members, m)
		}
	}
	return members, nil
}

/*
	{
		"id": 4399001234,
		"matchId": 4399001234,
		"userId": 0,
		"displayIdentity": "KM1234G1",
		"venue": "",
		"location": "",
		"matchScoreAdded": true,
		"league": "Boston Pickleball Classic powered by World Pickleball Tour - $5K Cash Purse - Women's Doubles 3.5",
		"eventDate": "2022-121-01",
		"eventFormat": "DOUBLES",
		"scoreFormat": {
		   "id": 5994912345,
		   "format": "Best 2 out of 3 Games to 11",
		   "games": 3,
		   "winningScore": 11
		},
		"confirmed": true,
		"teams": [
		   {
			  "id": 6280112345,
			  "serial": 1,
			  "player1": {
				 "id": 6111512345,
				 "fullName": "Mary Smith",
				 "duprId": "12AV70",
				 "imageUrl": null
			  },
			  "player2": {
				 "id": 8315312345,
				 "fullName": "Kim Jones",
				 "duprId": "12QX32",
				 "imageUrl": null
			  },
			  "game1": 11,
			  "game2": 11,
			  "game3": -1,
			  "game4": -1,
			  "game5": -1,
			  "winner": true,
			  "delta": "",
			  "teamRating": ""
		   },
		   {
			  "id": 5111412345,
			  "serial": 2,
			  "player1": {
				 "id": 5981212345,
				 "fullName": "Pam Brown",
				 "duprId": "323QW1",
				 "imageUrl": null
			  },
			  "player2": {
				 "id": 6655012345,
				 "fullName": "Kate Jones",
				 "duprId": "332AB1",
				 "imageUrl": null
			  },
			  "game1": 5,
			  "game2": 7,
			  "game3": -1,
			  "game4": -1,
			  "game5": -1,
			  "winner": false,
			  "delta": "",
			  "teamRating": ""
		   }
		],
		"created": "2022-08-23T16:06:36.114168Z",
		"eventName": "Boston Pickleball Classic powered by World Pickleball Tour - $5K Cash Purse - Women's Doubles 3.5",
		"matchSource": "MANUAL",
		"noOfGames": 2,
		"status": "ACTIVE"
	 }
*/
type APITeamPlayer struct {
	Id       uint64 `json:"id"`
	FullName string `json:"fullName"`
	DUPRId   string `json:"duprId"`
}

type APITeam struct {
	Id         uint64        `json:"id"`
	Player1    APITeamPlayer `json:"player1"`
	Player2    APITeamPlayer `json:"player2"`
	Game1      uint
	Game2      uint
	Game3      uint
	Game4      uint
	Game5      uint
	Winner     bool
	Delta      string
	TeamRating string
}

type APIMatch struct {
	id          uint64 `json:"id"`
	MatchId     uint64 `json:"matchId"`
	Venue       string `json:"venue"`
	Location    string `json:"location"`
	League      string `json:"league"`
	EventDate   string `json:"eventDate"`
	EventFormat string `json:"eventFormat"`
	EventName   string
	NoOfGames   uint
	Teams       [2]APITeam `json:"teams"`
}

func GetMemberMatchHistory(memberId string) ([]APIMatch, error) {
	// /club/{club_id}/members/v1.0/all
	log.Printf("GetMemberMachHistory %s\n", memberId)
	var pageIn MemberMatchHistoryPageRequest
	pageIn.Query = "*"
	pageIn.Limit = 10
	pageIn.Sort.Order = "DESC"
	pageIn.Sort.Parameter = "MATCH_DATE"

	log.Println(pageIn)
	dupr.LoadTokens()
	//goodJson := "{'exclude': [], 'limit': 20, 'offset': 0, 'query': '*'}"
	var result PagedResponse
	url := dupr.build_url("player/v1.0/" + memberId + "/history")

	var offset int = 0
	var matches []APIMatch

	for offset >= 0 {
		pageIn.Offset = uint(offset)
		_, err := dupr.GetRequest().SetBody(&pageIn).SetResult(&result).Post(url)
		if err != nil {
			log.Println("error")
			return nil, nil
		}
		offset = handlePaging(&result.Result)
		log.Println(len(result.Result.Hits))

		if offset >= 20 {
			return matches, nil
		}
		log.Println(result.Result.Hits[0])
		for _, h := range result.Result.Hits {
			var m APIMatch
			d, _ := json.Marshal(h)
			json.Unmarshal(d, &m)
			matches = append(matches, m)
		}
	}
	return matches, nil
}
