package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(host, port, user, password, dbname string) (*Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := initDB(db); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) AddTrackedPlayer(player *TrackedPlayer) error {
	query := `
		INSERT INTO tracked_players (puuid, game_name, tag_line, summoner_id, last_match_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (puuid) DO UPDATE SET
			game_name = $2, tag_line = $3, summoner_id = $4, last_match_id = $5, updated_at = $6`

	_, err := d.db.Exec(query, player.PUUID, player.GameName, player.TagLine, player.SummonerID, player.LastMatchID, time.Now())
	return err
}

func (d *Database) GetTrackedPlayers() ([]TrackedPlayer, error) {
	query := `SELECT id, puuid, game_name, tag_line, summoner_id, last_match_id, created_at, updated_at FROM tracked_players`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []TrackedPlayer
	for rows.Next() {
		var player TrackedPlayer
		err := rows.Scan(&player.ID, &player.PUUID, &player.GameName, &player.TagLine,
			&player.SummonerID, &player.LastMatchID, &player.CreatedAt, &player.UpdatedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (d *Database) RemoveTrackedPlayer(puuid string) error {
	query := `DELETE FROM tracked_players WHERE puuid = $1`
	_, err := d.db.Exec(query, puuid)
	return err
}

func (d *Database) UpdateLastMatchID(puuid, matchID string) error {
	query := `UPDATE tracked_players SET last_match_id = $1, updated_at = $2 WHERE puuid = $3`
	_, err := d.db.Exec(query, matchID, time.Now(), puuid)
	return err
}

func (d *Database) AddMatchData(match *MatchData) error {
	query := `
		INSERT INTO match_data 
		(match_id, puuid, champion, game_mode, game_duration, win, kills, deaths, assists, 
		 creep_score, damage_dealt, damage_taken, vision_score, gold_earned, items, game_creation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (match_id, puuid) DO NOTHING`

	_, err := d.db.Exec(query, match.MatchID, match.PUUID, match.Champion, match.GameMode,
		match.GameDuration, match.Win, match.Kills, match.Deaths, match.Assists,
		match.CreepScore, match.DamageDealt, match.DamageTaken, match.VisionScore,
		match.GoldEarned, match.Items, match.GameCreation)
	return err
}

func (d *Database) GetPlayerStats(puuid string, days int) ([]MatchData, error) {
	query := `
		SELECT match_id, puuid, champion, game_mode, game_duration, win, kills, deaths, assists,
		       creep_score, damage_dealt, damage_taken, vision_score, gold_earned, items, game_creation, extracted_at
		FROM match_data 
		WHERE puuid = $1 AND game_creation >= NOW() - INTERVAL '%d days'
		ORDER BY game_creation DESC`

	formattedQuery := fmt.Sprintf(query, days)
	rows, err := d.db.Query(formattedQuery, puuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []MatchData
	for rows.Next() {
		var match MatchData
		err := rows.Scan(&match.MatchID, &match.PUUID, &match.Champion, &match.GameMode,
			&match.GameDuration, &match.Win, &match.Kills, &match.Deaths, &match.Assists,
			&match.CreepScore, &match.DamageDealt, &match.DamageTaken, &match.VisionScore,
			&match.GoldEarned, &match.Items, &match.GameCreation, &match.ExtractedAt)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (d *Database) GetPlayerByRiotID(gameName, tagLine string) (*TrackedPlayer, error) {
	query := `SELECT id, puuid, game_name, tag_line, summoner_id, last_match_id, created_at, updated_at 
			  FROM tracked_players WHERE game_name = $1 AND tag_line = $2`

	var player TrackedPlayer
	err := d.db.QueryRow(query, gameName, tagLine).Scan(
		&player.ID, &player.PUUID, &player.GameName, &player.TagLine,
		&player.SummonerID, &player.LastMatchID, &player.CreatedAt, &player.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &player, nil
}
