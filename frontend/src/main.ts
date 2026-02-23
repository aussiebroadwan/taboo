import "./style.css";
import { SSEClient } from "./sse";
import { usingDiscordSDK } from "./discord";
import { mountLiveDraw } from "./views/live-draw";
import { mountPreviousDraw } from "./views/previous-draw";
import logger from "./logger";

const log = logger.with({ component: "main" });

function getSSEUrl(): string {
    const protocol = window.location.protocol;
    const hostname = window.location.host;
    const baseUrl = `${protocol}//${hostname}${usingDiscordSDK ? "/.proxy" : ""}`;
    return `${baseUrl}/api/v1/events`;
}

function route(app: HTMLElement): void {
    const path = window.location.pathname;
    const gamePathMatch = path.match(/^\/game\/(\d+)$/);

    if (gamePathMatch) {
        const gameId = Number(gamePathMatch[1]);
        if (isNaN(gameId)) {
            log.error("Invalid game ID", { path });
            window.location.href = "/";
            return;
        }

        mountPreviousDraw(app, gameId);
    } else if (path === "/") {
        const sseClient = new SSEClient(getSSEUrl(), { reconnectInterval: 3000 });
        mountLiveDraw(app, sseClient);
        sseClient.connect();
    } else {
        log.warn("Page not found, redirecting to live", { path });
        window.location.href = "/";
    }
}

document.addEventListener("DOMContentLoaded", () => {
    const app = document.getElementById("app");
    if (!app) {
        log.error("App element not found", { element_id: "app" });
        return;
    }

    route(app);
});
