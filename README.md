# League of Legends Discord Bot

A Discord bot that tracks League of Legends players and automatically posts game summaries when they finish matches.

## Features

- **Player Tracking**: Track specific League of Legends players
- **Automatic Game Detection**: Monitors for new games every 5 minutes
- **Rich Game Summaries**: Detailed match information including KDA, CS, damage, and more
- **Player Statistics**: View aggregated stats for tracked players
- **Discord Integration**: Full slash command support
- **Database Storage**: PostgreSQL database for scalable player and match data storage
- **Containerized**: Easy deployment with Docker

## Setup Instructions

### 1. Create Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to "Bot" section and create a bot
4. Copy the bot token
5. Under "Privileged Gateway Intents", enable "Message Content Intent"

### 2. Invite Bot to Server

1. Go to "OAuth2" > "URL Generator"
2. Select "bot" scope
3. Select "Send Messages" and "Read Message History" permissions
4. Use the generated URL to invite the bot to your server

### 2. Get Riot API Key

1. Go to [Riot Developer Portal](https://developer.riotgames.com/)
2. Sign in with your Riot account
3. Create a new app and get your API key
4. **Important**: Development keys expire every 24 hours. For production use, apply for a production key.

### 3. Deploy on NAS

#### Option A: Using Docker Compose (Recommended)

1. Copy all files to your NAS
2. Create a `.env` file with your credentials:
   ```
   DISCORD_TOKEN=your_bot_token_here
   RIOT_API_KEY=your_riot_api_key_here
   MONITOR_CHANNEL_ID=your_discord_channel_id_here
   
   # PostgreSQL Configuration (optional - uses defaults if not specified)
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=lol_bot
   ```
3. Run the deployment:
   ```bash
   docker-compose up -d
   ```
   This will start:
   - PostgreSQL database container
   - Discord bot application
   - Prometheus monitoring
   - Grafana dashboard

#### Option B: Using Docker Run

1. Build the image:
   ```bash
   docker build -t discord-bot .
   ```
2. Start PostgreSQL:
   ```bash
   docker run -d --name postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=lol_bot \
     -p 5432:5432 \
     -v postgres_data:/var/lib/postgresql/data \
     postgres:15-alpine
   ```
3. Run the bot container:
   ```bash
   docker run -d --name discord-bot \
     -e DISCORD_TOKEN=your_bot_token_here \
     -e RIOT_API_KEY=your_riot_api_key_here \
     -e MONITOR_CHANNEL_ID=your_discord_channel_id_here \
     -e DB_HOST=localhost \
     -e DB_PORT=5432 \
     -e DB_USER=postgres \
     -e DB_PASSWORD=postgres \
     -e DB_NAME=lol_bot \
     --network host \
     discord-bot
   ```

## Usage

### Discord Commands

The bot supports the following slash commands:

- `/track <summoner>` - Track a League of Legends player (e.g., `/track PlayerName#TAG`)
- `/untrack <summoner>` - Stop tracking a player
- `/stats <summoner> [days]` - Show player statistics (default: 7 days)
- `/tracked` - List all currently tracked players
- `/pn` or `/patchnotes` - Get latest League of Legends patch notes
- `/help` - Show command help

### Automatic Game Monitoring

Once you track players, the bot will:
- Check for new games every 5 minutes
- Post detailed game summaries to the specified Discord channel
- Store match data in the database for statistics
- Track KDA, CS, damage, vision score, and more

### Game Summary Features

Each game summary includes:
- **Win/Loss status** with colored indicators
- **Champion played** and **KDA ratio**
- **CS (Creep Score)** and **damage dealt**
- **Vision score** and **gold earned**
- **Game mode** and **match duration**
- **Match ID** for reference

## Managing the Bot

- **View logs**: `docker-compose logs discord-bot`
- **Stop bot**: `docker-compose down`
- **Restart bot**: `docker-compose restart`
- **Update bot**: `docker-compose pull && docker-compose up -d`

## Troubleshooting

- Ensure your Discord token is correct
- Check that the bot has proper permissions in your Discord server
- Verify the bot is online in your Discord server member list
- Check container logs for any error messages

## File Structure

```
discord-bot/
├── main.go              # Main bot implementation and handlers
├── models.go            # Database models and PostgreSQL schema
├── riot_api.go          # Riot API client and data structures
├── database.go          # PostgreSQL database operations and queries
├── game_monitor.go      # Background game monitoring service
├── go.mod               # Go dependencies (discordgo, lib/pq, cron)
├── go.sum               # Go module checksums
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Deployment with PostgreSQL, Prometheus, Grafana
├── .env.example         # Environment template with DB config
├── prometheus.yml       # Prometheus monitoring configuration
└── README.md           # This file
```

## Environment Variables

### Required
- `DISCORD_TOKEN` - Your Discord bot token
- `RIOT_API_KEY` - Your Riot Games API key

### Optional
- `MONITOR_CHANNEL_ID` - Discord channel ID for game summaries
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL username (default: postgres)
- `DB_PASSWORD` - PostgreSQL password (default: postgres)
- `DB_NAME` - PostgreSQL database name (default: lol_bot)

## Database Schema

The bot uses PostgreSQL with two main tables:

### tracked_players
```sql
CREATE TABLE tracked_players (
    id SERIAL PRIMARY KEY,
    puuid VARCHAR(78) UNIQUE NOT NULL,
    game_name VARCHAR(255) NOT NULL,
    tag_line VARCHAR(16) NOT NULL,
    summoner_id VARCHAR(63) NOT NULL,
    last_match_id VARCHAR(32),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### match_data
```sql
CREATE TABLE match_data (
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
);
```

## API Rate Limits

The bot respects Riot API rate limits:
- Personal API keys: 100 requests every 2 minutes
- Production keys: Higher limits available
- The bot makes minimal API calls (only when checking for new games)

## Troubleshooting

- **"Player not found"**: Ensure the summoner name format is correct (PlayerName#TAG)
- **API key expired**: Development keys expire every 24 hours - get a new one from Riot Developer Portal
- **No game summaries**: Check that `MONITOR_CHANNEL_ID` is set and the bot has permissions to post in that channel
- **Database connection failed**: Check PostgreSQL container is running and environment variables are correct
- **Database migration errors**: Ensure PostgreSQL user has CREATE TABLE permissions