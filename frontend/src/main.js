import { Engine } from "./engine";
import { LiveDrawScene } from "./scenes/LiveDrawScene";

import Client from './network/Client.js';
import { DESIGN_WIDTH, DESIGN_HEIGHT } from './constants';

import { DiscordSDK } from "@discord/embedded-app-sdk";

let usingDiscordSDK = false;

// Check if the URL has the "frame_id" query parameter.
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.has("frame_id")) {
    usingDiscordSDK = true;

    // Instantiate and set up the Discord SDK.
    const discordSdk = new DiscordSDK(import.meta.env.VITE_DISCORD_CLIENT_ID);
    discordSdk.ready().then(() => console.log("Discord SDK is ready"));
}

const setupWebsocket = (scene) => {
    // Build the WebSocket URL based on the frontend's location.
    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const hostname = window.location.host;
    const wsUrl = `${protocol}${hostname}${usingDiscordSDK ? '/.proxy' : ''}/ws`;

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

    const liveDrawScene = new LiveDrawScene(engine);
    setupWebsocket(liveDrawScene);

    engine.registerScene("live-draw", liveDrawScene);
    engine.setScene("live-draw");

    // Start the engine's update and render loop.
    engine.start();
});
