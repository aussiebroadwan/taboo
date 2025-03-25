import { Engine } from "./engine";
import { LiveDrawScene } from "./scenes/LiveDrawScene";
import { PreviousDrawScene } from "./scenes/PreviousDrawScene";

import Client from './network/Client.js';
import { DESIGN_WIDTH, DESIGN_HEIGHT } from './constants';

import { DiscordSDK } from "@discord/embedded-app-sdk";

const urlParams = new URLSearchParams(window.location.search);
const usingDiscordSDK = urlParams.has("frame_id");

const path = window.location.pathname;
const protocol = window.location.protocol;
const hostname = window.location.host;

if (usingDiscordSDK) {
    fetch(`${protocol}//${hostname}/.proxy/client-id`)
        .then(response => response.json())
        .then(data => {
            // Instantiate and set up the Discord SDK.
            const discordSdk = new DiscordSDK(data.clientId);
            discordSdk.ready().then(() => console.log("Discord SDK is ready"));
        })
        .catch(error => console.error('Error fetching client ID:', error));
}

const setupWebsocket = (scene) => {
    // Build the WebSocket URL based on the frontend's location.
    const wsProtocol = protocol === 'https:' ? 'wss://' : 'ws://';
    const wsUrl = `${wsProtocol}${hostname}${usingDiscordSDK ? '/.proxy' : ''}/ws`;

    // Prepare Connection
    const wsClient = new Client(wsUrl, { reconnectInterval: 3000 });
    wsClient.onMessage = scene.onWebsocketMessage;
    wsClient.connect();
}

function route(engine) {
    const pathRe = /\/game\/\d+/g;

    if (path.match(pathRe)) {
        // Extract game id if needed.
        const segments = path.split('/');
        const gameIdStr = segments[2];

        // Convert the gameId string to a number and verify it's numeric.
        const gameId = Number(gameIdStr);
        if (isNaN(gameId)) {
            console.error("router: invalid game id");
            window.location.href = '/';
            return;
        }

        // Create and register the PreviousDrawScene for the specified game.
        const previousDrawScene = new PreviousDrawScene(engine, gameId, usingDiscordSDK);
        engine.registerScene("game", previousDrawScene);
        engine.setScene("game");

    } else if (path === "/") {

        // At the homepage, use the LiveDrawScene.
        const liveDrawScene = new LiveDrawScene(engine);
        setupWebsocket(liveDrawScene);

        engine.registerScene("live-draw", liveDrawScene);
        engine.setScene("live-draw");

    } else {
        // For any unknown path, log the error and redirect to the homepage.
        console.log("router: page not found routing to live");
        window.location.href = '/';
        return;
    }
}

/* Entry Point */
document.addEventListener('DOMContentLoaded', () => {
    const canvas = document.getElementById('canvas');
    if (!canvas) {
        console.error('Canvas element with id "canvas" not found.');
        return;
    }

    // Create the engine instance. Here our design dimensions are 720x364.
    const engine = new Engine(canvas, DESIGN_WIDTH, DESIGN_HEIGHT);

    // Execute routing logic to initialise the appropriate scene.
    route(engine);

    // Start the engine's update and render loop.
    engine.start();
});
