import logger from "./logger";

const log = logger.with({ component: "discord_sdk" });

const urlParams = new URLSearchParams(window.location.search);
export const usingDiscordSDK = urlParams.has("frame_id");

interface DiscordSDKInstance {
    commands: {
        authorize(opts: {
            client_id: string;
            response_type: string;
            state: string;
            prompt: string;
            scope: string[];
        }): Promise<{ code: string }>;
        authenticate(opts: { access_token: string }): Promise<unknown>;
        setActivity(opts: { activity: unknown }): Promise<void>;
    };
    ready(): Promise<void>;
}

let sdkInitPromise: Promise<DiscordSDKInstance | null> | null = null;

const scopes = [
    "identify",
    "rpc.activities.write",
];

async function initDiscordSDK(): Promise<DiscordSDKInstance | null> {
    if (!usingDiscordSDK) return null;
    if (sdkInitPromise) return sdkInitPromise;

    sdkInitPromise = (async () => {
        const { DiscordSDK } = await import("@discord/embedded-app-sdk");

        const protocol = window.location.protocol;
        const hostname = window.location.hostname;

        const clientIdResponse = await fetch(`${protocol}//${hostname}/.proxy/client-id`);
        const { clientId } = await clientIdResponse.json();

        const sdk = new DiscordSDK(clientId) as unknown as DiscordSDKInstance;
        await sdk.ready();
        log.info("SDK ready");

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

        return sdk;
    })();

    return sdkInitPromise;
}

/**
 * Runs a callback with the Discord SDK when it's ready.
 * No-ops if not running inside Discord.
 */
export async function useDiscordSDK(
    callback: (sdk: DiscordSDKInstance) => void | Promise<void>,
): Promise<void> {
    if (!usingDiscordSDK) return;

    const sdk = await initDiscordSDK();
    if (sdk) {
        await callback(sdk);
    }
}
