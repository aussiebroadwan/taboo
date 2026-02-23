import { COLORS } from "../constants";
import { createGrid } from "../components/grid";
import { createLargeCounter, createSmallCounter, type Counter } from "../components/counter";
import { placePickInstant, placePickAnimated } from "../components/pick";
import { createGameState, resetGameState, addPick, type GameState } from "../state";
import { useDiscordSDK } from "../discord";
import type { SSEClient } from "../sse";
import type { GameStateData, GamePickData } from "../types";
import logger from "../logger";

const log = logger.with({ component: "live_draw_view" });

const MS_PER_MINUTE = 60_000;
const MS_PER_SECOND = 1_000;

const GAME_DRAW_TIME = 1.5 * MS_PER_MINUTE;
const GAME_WAIT_TIME = 1.5 * MS_PER_MINUTE;
const GAME_TOTAL_TIME = GAME_DRAW_TIME + GAME_WAIT_TIME;

export function mountLiveDraw(root: HTMLElement, sseClient: SSEClient): void {
    const container = document.createElement("div");
    container.className = "game-container";
    root.appendChild(container);

    // Build UI components
    const { element: gridEl, cells } = createGrid();
    const timerCounter = createLargeCounter("NEXT GAME", COLORS.BLUE);
    const drawCounter = createLargeCounter("DRAWING GAME", COLORS.BLUE);
    const headsCounter = createSmallCounter("Heads", COLORS.RED);
    const tailsCounter = createSmallCounter("Tails", COLORS.BLUE);

    // Counter bar
    const counterBar = document.createElement("div");
    counterBar.className = "counter-bar";
    counterBar.appendChild(headsCounter.element);
    counterBar.appendChild(tailsCounter.element);
    counterBar.appendChild(timerCounter.element);
    counterBar.appendChild(drawCounter.element);

    container.appendChild(counterBar);
    container.appendChild(gridEl);

    // Loading overlay (hidden once first game:state arrives)
    const overlay = document.createElement("div");
    overlay.className = "loading-overlay";
    const spinner = document.createElement("div");
    spinner.className = "loading-spinner";
    const label = document.createElement("div");
    label.className = "loading-label";
    label.textContent = "Loading...";
    overlay.appendChild(spinner);
    overlay.appendChild(label);
    container.appendChild(overlay);

    // State
    const state = createGameState();

    // Timer
    const timerInterval = setInterval(() => {
        timerCounter.setValue(getTimeLeftString(state));
    }, MS_PER_SECOND);

    // SSE handlers
    sseClient.onGameState = (data: GameStateData) => {
        handleGameState(
            container, cells, state, data,
            timerCounter, drawCounter, headsCounter, tailsCounter,
        );
    };

    sseClient.onGamePick = (data: GamePickData) => {
        handleGamePick(container, cells, state, data, headsCounter, tailsCounter);
    };

    sseClient.onGameComplete = (data) => {
        log.info("Game completed", { game_id: data.game_id });
    };

    // Store cleanup reference on the container for potential future use
    container.dataset.timerInterval = String(timerInterval);
}

function handleGameState(
    container: HTMLElement,
    cells: Map<number, HTMLElement>,
    state: GameState,
    data: GameStateData,
    timerCounter: Counter,
    drawCounter: Counter,
    headsCounter: Counter,
    tailsCounter: Counter,
): void {
    // Dismiss loading overlay on first game:state
    const overlay = container.querySelector(".loading-overlay");
    if (overlay) {
        overlay.classList.add("loading-overlay--hidden");
        overlay.addEventListener("transitionend", () => overlay.remove(), { once: true });
    }

    const isNewGame = data.game_id !== state.gameId;

    state.nextGame = data.next_game;

    if (isNewGame) {
        state.gameId = data.game_id;
        resetGameState(state);

        // Clear all cell colors
        cells.forEach((cell) => {
            cell.style.backgroundColor = "#D9D9D9";
            cell.classList.remove("grid-cell--picked");
        });

        // Remove any lingering flying picks
        container.querySelectorAll(".pick-fly").forEach((el) => el.remove());

        // Place existing picks instantly
        for (const pick of data.picks) {
            addPick(state, pick);
            placePickInstant(cells, pick);
        }

        drawCounter.setValue(state.gameId);
        headsCounter.setValue(state.heads);
        tailsCounter.setValue(state.tails);
        timerCounter.setValue(getTimeLeftString(state));

        // Set Discord rich presence
        const gameStart = new Date(state.nextGame).getTime() - GAME_TOTAL_TIME;
        useDiscordSDK((sdk) =>
            sdk.commands.setActivity({
                activity: {
                    type: 3, // Watching
                    details: `Game ${state.gameId}`,
                    timestamps: {
                        start: gameStart,
                        end: new Date(state.nextGame).getTime(),
                    },
                },
            }).then(() => log.debug("Activity rich presence set", { game_id: state.gameId })),
        );
    }
}

function handleGamePick(
    container: HTMLElement,
    cells: Map<number, HTMLElement>,
    state: GameState,
    data: GamePickData,
    headsCounter: Counter,
    tailsCounter: Counter,
): void {
    addPick(state, data.pick);
    placePickAnimated(container, cells, data.pick);
    headsCounter.setValue(state.heads);
    tailsCounter.setValue(state.tails);
}

function getTimeLeftString(state: GameState): string {
    if (!state.nextGame) return "00:00";
    const now = Date.now();
    const timeLeft = new Date(state.nextGame).getTime() - now;
    if (timeLeft < 0) return "00:00";

    const minutes = Math.floor(timeLeft / MS_PER_MINUTE);
    const seconds = Math.floor((timeLeft - minutes * MS_PER_MINUTE) / MS_PER_SECOND);

    return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}
