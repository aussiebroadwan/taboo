import { COLORS } from "../constants";
import { createGrid } from "../components/grid";
import { createLargeCounter, createSmallCounter } from "../components/counter";
import { placePickInstant } from "../components/pick";
import { createGameState, addPick } from "../state";
import { usingDiscordSDK } from "../discord";
import type { GameResponse } from "../types";
import logger from "../logger";

const log = logger.with({ component: "previous_draw_view" });

export function mountPreviousDraw(root: HTMLElement, gameId: number): void {
    const container = document.createElement("div");
    container.className = "game-container";
    root.appendChild(container);

    // Build UI components
    const { element: gridEl, cells } = createGrid();
    const timerCounter = createLargeCounter("NEXT GAME", COLORS.BLUE);
    const drawCounter = createLargeCounter("DRAWING GAME", COLORS.BLUE);
    const headsCounter = createSmallCounter("Heads", COLORS.RED);
    const tailsCounter = createSmallCounter("Tails", COLORS.BLUE);

    timerCounter.setValue("00:00");
    drawCounter.setValue(gameId);

    // Counter bar
    const counterBar = document.createElement("div");
    counterBar.className = "counter-bar";
    counterBar.appendChild(headsCounter.element);
    counterBar.appendChild(tailsCounter.element);
    counterBar.appendChild(timerCounter.element);
    counterBar.appendChild(drawCounter.element);

    container.appendChild(counterBar);
    container.appendChild(gridEl);

    // Fetch game data
    const protocol = window.location.protocol;
    const hostname = window.location.host;
    const baseUrl = `${protocol}//${hostname}${usingDiscordSDK ? "/.proxy" : ""}`;
    const apiUrl = `${baseUrl}/api/v1/games/${gameId}`;

    fetch(apiUrl)
        .then((response) => {
            if (response.status !== 200) {
                response.text().then((body) => {
                    log.warn("Failed to fetch draw, redirecting to live", {
                        game_id: gameId,
                        status: response.status,
                        body,
                    });
                    window.location.href = "/";
                });
                return;
            }

            return response.json().then((data: GameResponse) => {
                const state = createGameState();
                state.gameId = gameId;

                for (const pick of data.picks) {
                    addPick(state, pick);
                    placePickInstant(cells, pick);
                }

                headsCounter.setValue(state.heads);
                tailsCounter.setValue(state.tails);

                log.info("Previous game drawn", {
                    game_id: gameId,
                    picks: data.picks.length,
                });
            });
        })
        .catch((error: unknown) => {
            log.error("Failed to fetch game, redirecting to live", {
                game_id: gameId,
                error: String(error),
            });
            window.location.href = "/";
        });
}
