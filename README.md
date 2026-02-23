# TAB - Taboo

As typical Australian Developers, pubs have been a big part of our lives and
shape us into the unruly people we are today. A big portion of our time in
pubs are spent drinking, horse betting and losing a house deposit worth of
money on Keno. Keno brings more joy to our remaining brain cells than it
probably should, so to quench our insatiable need to watch the numbers 1 to
80 flash on a screen we have built a Keno-lite dubbed Taboo. The live web
view is available at https://taboo.tabdiscord.com.

## What is Keno?

Keno is a lottery-style game of chance where players pick numbers from a fixed
range (usually 1 to 80), and then a set of numbers is drawn at random. The
objective is to match as many of your chosen numbers as possible with the
drawn numbers to win prizes.

> **Note**: This project is only the graphical and functional part of the
>           lottery draw. This doesn't handle betting or accounts, this is
>           only for background entertainment.

## Discord Activities & Web Viewing

Taboo is built to offer a seamless gaming experience as both a web
application and a Discord Activity.

- **Discord Activities**:
    Taboo’s Discord implementation is designed for group viewing during
    video/voice calls, adding a light-hearted, fun element to drinking sessions
    and social hangouts.

- **General Web Viewing**:
    For those not on Discord, Taboo functions as a standard web app, delivering
    the same real-time Keno experience.

## Deployment

### Docker

A Discord app is required to enable the Discord Activity integration. You must
provide the Discord client secret and client id as environment variables for
Taboo to function properly. For guidance on setting up your Discord app, refer
to the [Discord Developer Docs].

```yaml
services:
  taboo:
    image: ghcr.io/aussiebroadwan/taboo:latest
    ports:
      - "8080:8080"
    environment:
      - DISCORD_CLIENT_SECRET=<your-secret-here>
      - DISCORD_CLIENT_ID=<your-client-id-here>
    restart: always
    volumes:
      - ./data:/data
```

## How It Functions

Taboo operates by continuously generating random numbers using its dedicated
game engine, which updates the game state in real time. Live updates are
streamed from the Go backend to the embedded frontend via Server-Sent Events
(SSE), ensuring that clients always receive the latest game information. The
application compiles to a single binary with all static assets embedded,
making deployment straightforward.

- **Live Game Rounds**: Continuous generation and broadcast of random picks
    creates a dynamic gaming experience.

- **Real-time Streaming**: Server-Sent Events (SSE) deliver structured updates
    to all connected clients with automatic reconnection support.

## Support & Contribution

Please note that support and development for Taboo will be inconsistent. We
welcome any help with additional features or improvements—community
contributions are always appreciated!


[Discord Developer Docs]: https://discord.com/developers/docs/quick-start/getting-started
