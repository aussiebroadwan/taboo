
let sdkInstance = null;
let sdkReady = false;
let sdkInitPromise = null;

// Change this when you need different scopes for the DiscordSDK
const scopes = [
    "identify",
    "rpc.activities.write" // Allow `sdk.commands.setActivities()``
];

const urlParams = new URLSearchParams(window.location.search);
const usingDiscordSDK = urlParams.has("frame_id");

async function initDiscordSDK() {
    if (!usingDiscordSDK) return null;
    if (sdkInitPromise) return sdkInitPromise;

    sdkInitPromise = (async () => {
        const { DiscordSDK } = await import("@discord/embedded-app-sdk");

        const protocol = window.location.protocol;
        const hostname = window.location.hostname;

        const clientIdResponse = await fetch(`${protocol}//${hostname}/.proxy/client-id`);
        const { clientId } = await clientIdResponse.json();

        const sdk = new DiscordSDK(clientId);
        await sdk.ready();
        console.log("discordsdk: sdk is ready");

        const { code } = await sdk.commands.authorize({
            client_id: clientId,
            response_type: "code",
            state: "",
            prompt: "none",
            scope: scopes,
        });

        const tokenResponse = await fetch(`${protocol}//${hostname}/.proxy/api/token`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ code }),
        });

        const { access_token } = await tokenResponse.json();

        const auth = await sdk.commands.authenticate({ access_token });
        if (!auth) throw new Error("Authentication with Discord SDK failed.");

        sdkInstance = sdk;
        sdkReady = true;
        return sdkInstance;
    })();

    return sdkInitPromise;
}

/**
 * Asynchronously uses the Discord SDK when it's ready.
 * 
 * @param {(sdk: any) => void | Promise<void>} callback 
 * @returns {Promise<void>}
 */
async function useDiscordSDK(callback) {
    if (!usingDiscordSDK) return;

    const sdk = await initDiscordSDK();
    if (sdk) {
        // Always call callback asynchronously
        // await Promise.resolve();
        await callback(sdk);
    }
}

export { useDiscordSDK, usingDiscordSDK };
