/**
 * Game state helpers for tracking picks and counters.
 */

export interface GameState {
    gameId: number;
    picks: number[];
    heads: number;
    tails: number;
    nextGame: string;
}

export function createGameState(): GameState {
    return {
        gameId: 0,
        picks: [],
        heads: 0,
        tails: 0,
        nextGame: "",
    };
}

export function resetGameState(state: GameState): void {
    state.picks = [];
    state.heads = 0;
    state.tails = 0;
}

export function addPick(state: GameState, pick: number): void {
    state.picks.push(pick);
    if (pick > 40) {
        state.tails++;
    } else {
        state.heads++;
    }
}
