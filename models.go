package main

import (
	"database/sql"
	"time"
)

type TrackedPlayer struct {
	ID          int       `db:"id"`
	PUUID       string    `db:"puuid"`
	GameName    string    `db:"game_name"`
	TagLine     string    `db:"tag_line"`
	SummonerID  string    `db:"summoner_id"`
	LastMatchID string    `db:"last_match_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type MatchData struct {
	ID             int       `db:"id"`
	MatchID        string    `db:"match_id"`
	PUUID          string    `db:"puuid"`
	Champion       string    `db:"champion"`
	GameMode       string    `db:"game_mode"`
	GameDuration   int       `db:"game_duration"`
	Win            bool      `db:"win"`
	Kills          int       `db:"kills"`
	Deaths         int       `db:"deaths"`
	Assists        int       `db:"assists"`
	CreepScore     int       `db:"creep_score"`
	DamageDealt    int       `db:"damage_dealt"`
	DamageTaken    int       `db:"damage_taken"`
	VisionScore    int       `db:"vision_score"`
	GoldEarned     int       `db:"gold_earned"`
	Items          string    `db:"items"` // JSON string
	GameCreation   time.Time `db:"game_creation"`
	ExtractedAt    time.Time `db:"extracted_at"`
}

func initDB(db *sql.DB) error {
	createPlayersTable := `
	CREATE TABLE IF NOT EXISTS tracked_players (
		id SERIAL PRIMARY KEY,
		puuid VARCHAR(78) UNIQUE NOT NULL,
		game_name VARCHAR(255) NOT NULL,
		tag_line VARCHAR(16) NOT NULL,
		summoner_id VARCHAR(63) NOT NULL,
		last_match_id VARCHAR(32),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createMatchesTable := `
	CREATE TABLE IF NOT EXISTS match_data (
		id SERIAL PRIMARY KEY,
		match_id VARCHAR(32) NOT NULL,
		puuid VARCHAR(78) NOT NULL,
		champion VARCHAR(50) NOT NULL,
		game_mode VARCHAR(50) NOT NULL,
		game_duration INTEGER NOT NULL,
		win BOOLEAN NOT NULL,
		kills INTEGER NOT NULL,
		deaths INTEGER NOT NULL,
		assists INTEGER NOT NULL,
		creep_score INTEGER NOT NULL,
		damage_dealt INTEGER NOT NULL,
		damage_taken INTEGER NOT NULL,
		vision_score INTEGER NOT NULL,
		gold_earned INTEGER NOT NULL,
		items TEXT NOT NULL,
		game_creation TIMESTAMP NOT NULL,
		extracted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(match_id, puuid)
	);`

	if _, err := db.Exec(createPlayersTable); err != nil {
		return err
	}
	if _, err := db.Exec(createMatchesTable); err != nil {
		return err
	}

	return nil
}