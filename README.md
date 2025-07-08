# League of Legends Discord Bot

A Discord bot that tracks League of Legends players and automatically posts game summaries when they finish matches.

## Features

- **Player Tracking**: Track specific League of Legends players
- **Automatic Game Detection**: Monitors for new games every 5 minutes
- **Rich Game Summaries**: Detailed match information including KDA, CS, damage, and more
- **Player Statistics**: View aggregated stats for tracked players
- **Discord Integration**: Full slash command support
- **Database Storage**: Persistent SQLite database for player and match data
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
   ```
3. Run the deployment:
   ```bash
   docker-compose up -d
   ```

#### Option B: Using Docker Run

1. Build the image:
   ```bash
   docker build -t discord-bot .
   ```
2. Run the container:
   ```bash
   docker run -d --name discord-bot \
     -e DISCORD_TOKEN=your_bot_token_here \
     -e RIOT_API_KEY=your_riot_api_key_here \
     -e MONITOR_CHANNEL_ID=your_discord_channel_id_here \
     -v $(pwd)/bot.db:/app/bot.db \
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
├── models.go            # Database models and schema
├── riot_api.go          # Riot API client and data structures
├── database.go          # Database operations and queries
├── game_monitor.go      # Background game monitoring service
├── go.mod               # Go dependencies
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Deployment configuration
├── .env.example         # Environment template
├── bot.db               # SQLite database (created on first run)
└── README.md           # This file
```

## Environment Variables

- `DISCORD_TOKEN` - Required: Your Discord bot token
- `RIOT_API_KEY` - Required: Your Riot Games API key
- `MONITOR_CHANNEL_ID` - Optional: Discord channel ID for game summaries

## Database Schema

The bot uses SQLite with two main tables:
- `tracked_players` - Stores player information and tracking status
- `match_data` - Stores detailed match statistics for analysis

## API Rate Limits

The bot respects Riot API rate limits:
- Personal API keys: 100 requests every 2 minutes
- Production keys: Higher limits available
- The bot makes minimal API calls (only when checking for new games)

## Troubleshooting

- **"Player not found"**: Ensure the summoner name format is correct (PlayerName#TAG)
- **API key expired**: Development keys expire every 24 hours - get a new one from Riot Developer Portal
- **No game summaries**: Check that `MONITOR_CHANNEL_ID` is set and the bot has permissions to post in that channel
- **Database locked**: Ensure proper file permissions for the SQLite database file