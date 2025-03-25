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
    const hostname = window.location.host;
    const wsUrl = `${wsProtocol}${hostname}${usingDiscordSDK ? '/.proxy' : ''}/ws`;

    // Prepare Connection
    const wsClient = new Client(wsUrl, { reconnectInterval: 3000 });

    // Set the Callback Events
    wsClient.onOpen = () => console.log('Connected to server.');
    wsClient.onMessage = scene.onWebsocketMessage;
    wsClient.onClose = () => console.log('Disconnected from server.');
    wsClient.onError = (error) => console.error('WebSocket error:', error);

    // Establish the WebSocket connection.
    wsClient.connect();
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

    if (path.startsWith('/game/')) {
        // Extract game id if needed.
        const segments = path.split('/');
        const gameIdStr = segments[2] || null;

        if (!gameIdStr) return;

        // Convert the gameId string to a number and verify it's numeric.
        const gameId = Number(gameIdStr);
        if (isNaN(gameId)) {
            console.log("router: invalid game id");
            window.location.href = '/';
            return;
        }

        const previousDrawScene = new PreviousDrawScene(engine, gameId, usingDiscordSDK);
        engine.registerScene("game", previousDrawScene);
        engine.setScene("game");
    } else {

        const liveDrawScene = new LiveDrawScene(engine);
        setupWebsocket(liveDrawScene);

        engine.registerScene("live-draw", liveDrawScene);
        engine.setScene("live-draw");
    }


    // Start the engine's update and render loop.
    engine.start();
});
