import logger from "./logger";
import type { GameStateData, GamePickData, GameCompleteData } from "./types";

const log = logger.with({ component: "sse" });

export interface SSEClientOptions {
    reconnectInterval?: number;
}

export type GameStateHandler = (data: GameStateData) => void;
export type GamePickHandler = (data: GamePickData) => void;
export type GameCompleteHandler = (data: GameCompleteData) => void;

export class SSEClient {
    private url: string;
    private eventSource: EventSource | null = null;
    private reconnectInterval: number;
    private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
    private intentionalClose = false;

    onGameState: GameStateHandler | null = null;
    onGamePick: GamePickHandler | null = null;
    onGameComplete: GameCompleteHandler | null = null;

    constructor(url: string, options: SSEClientOptions = {}) {
        this.url = url;
        this.reconnectInterval = options.reconnectInterval ?? 3000;
    }

    connect(): void {
        this.intentionalClose = false;
        this.eventSource = new EventSource(this.url);

        this.eventSource.onopen = () => {
            log.info("Connected", { url: this.url });
        };

        this.eventSource.onerror = () => {
            log.error("Connection error", { url: this.url });
            if (this.eventSource?.readyState === EventSource.CLOSED) {
                this.handleDisconnect();
            }
        };

        this.eventSource.addEventListener("game:state", (e) => {
            const data: GameStateData = JSON.parse((e as MessageEvent).data);
            this.onGameState?.(data);
        });

        this.eventSource.addEventListener("game:pick", (e) => {
            const data: GamePickData = JSON.parse((e as MessageEvent).data);
            this.onGamePick?.(data);
        });

        this.eventSource.addEventListener("game:complete", (e) => {
            const data: GameCompleteData = JSON.parse((e as MessageEvent).data);
            this.onGameComplete?.(data);
        });

        this.eventSource.addEventListener("game:heartbeat", () => {
            // Keep-alive, no action needed
        });
    }

    private handleDisconnect(): void {
        log.warn("Disconnected", { url: this.url });
        if (!this.intentionalClose) {
            this.reconnectTimeoutId = setTimeout(() => this.connect(), this.reconnectInterval);
        }
    }

    disconnect(): void {
        this.intentionalClose = true;
        if (this.reconnectTimeoutId) clearTimeout(this.reconnectTimeoutId);
        if (this.eventSource) this.eventSource.close();
    }
}
