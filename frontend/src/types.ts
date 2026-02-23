// SSE event types (matching Go sdk/events.go)

export interface GameStateData {
    game_id: number;
    picks: number[];
    next_game: string;
}

export interface GamePickData {
    pick: number;
}

export interface GameCompleteData {
    game_id: number;
}

// REST API types (matching Go sdk/dto.go)

export interface GameResponse {
    id: number;
    picks: number[];
    created_at: string;
}

export interface GameListResponse {
    games: GameResponse[];
    next_cursor?: number;
}

export interface ErrorResponse {
    error: {
        code: string;
        message: string;
    };
}

// SSE event names
export const SSE_GAME_STATE = "game:state";
export const SSE_GAME_PICK = "game:pick";
export const SSE_GAME_COMPLETE = "game:complete";
export const SSE_GAME_HEARTBEAT = "game:heartbeat";
