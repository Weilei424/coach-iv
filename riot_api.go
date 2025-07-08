package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RiotAPI struct {
	APIKey string
	Client *http.Client
}

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type Summoner struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	PUUID         string `json:"puuid"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int    `json:"summonerLevel"`
}

type Match struct {
	Info struct {
		GameID       int64  `json:"gameId"`
		GameMode     string `json:"gameMode"`
		GameDuration int    `json:"gameDuration"`
		GameCreation int64  `json:"gameCreation"`
		Participants []struct {
			PUUID              string `json:"puuid"`
			ChampionName       string `json:"championName"`
			Win                bool   `json:"win"`
			Kills              int    `json:"kills"`
			Deaths             int    `json:"deaths"`
			Assists            int    `json:"assists"`
			TotalMinionsKilled int    `json:"totalMinionsKilled"`
			TotalDamageDealt   int    `json:"totalDamageDealtToChampions"`
			TotalDamageTaken   int    `json:"totalDamageTaken"`
			VisionScore        int    `json:"visionScore"`
			GoldEarned         int    `json:"goldEarned"`
			Item0              int    `json:"item0"`
			Item1              int    `json:"item1"`
			Item2              int    `json:"item2"`
			Item3              int    `json:"item3"`
			Item4              int    `json:"item4"`
			Item5              int    `json:"item5"`
			Item6              int    `json:"item6"`
		} `json:"participants"`
	} `json:"info"`
}

func NewRiotAPI(apiKey string) *RiotAPI {
	return &RiotAPI{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *RiotAPI) makeRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Riot-Token", r.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (r *RiotAPI) GetAccountByRiotID(gameName, tagLine string) (*Account, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s", gameName, tagLine)

	body, err := r.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var account Account
	if err := json.Unmarshal(body, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *RiotAPI) GetSummonerByPUUID(puuid string) (*Summoner, error) {
	url := fmt.Sprintf("https://na1.api.riotgames.com/lol/summoner/v4/summoners/by-puuid/%s", puuid)

	body, err := r.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var summoner Summoner
	if err := json.Unmarshal(body, &summoner); err != nil {
		return nil, err
	}

	return &summoner, nil
}

func (r *RiotAPI) GetMatchHistory(puuid string, count int) ([]string, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?count=%d", puuid, count)

	body, err := r.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var matchIDs []string
	if err := json.Unmarshal(body, &matchIDs); err != nil {
		return nil, err
	}

	return matchIDs, nil
}

func (r *RiotAPI) GetMatchDetails(matchID string) (*Match, error) {
	url := fmt.Sprintf("https://americas.api.riotgames.com/lol/match/v5/matches/%s", matchID)

	body, err := r.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var match Match
	if err := json.Unmarshal(body, &match); err != nil {
		return nil, err
	}

	return &match, nil
}

func (r *RiotAPI) ExtractPlayerData(match *Match, puuid string) *MatchData {
	for _, participant := range match.Info.Participants {
		if participant.PUUID == puuid {
			items := []int{
				participant.Item0, participant.Item1, participant.Item2,
				participant.Item3, participant.Item4, participant.Item5, participant.Item6,
			}
			itemsJSON, _ := json.Marshal(items)

			return &MatchData{
				MatchID:      fmt.Sprintf("%d", match.Info.GameID),
				PUUID:        puuid,
				Champion:     participant.ChampionName,
				GameMode:     match.Info.GameMode,
				GameDuration: match.Info.GameDuration,
				Win:          participant.Win,
				Kills:        participant.Kills,
				Deaths:       participant.Deaths,
				Assists:      participant.Assists,
				CreepScore:   participant.TotalMinionsKilled,
				DamageDealt:  participant.TotalDamageDealt,
				DamageTaken:  participant.TotalDamageTaken,
				VisionScore:  participant.VisionScore,
				GoldEarned:   participant.GoldEarned,
				Items:        string(itemsJSON),
				GameCreation: time.Unix(match.Info.GameCreation/1000, 0),
			}
		}
	}
	return nil
}
