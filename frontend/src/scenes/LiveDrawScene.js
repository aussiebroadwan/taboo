import { Scene } from '../engine';
import { LargeCounterComponent } from '../components/LargeCounterComponent';
import { SmallCounterComponent } from '../components/SmallCounterComponent';
import { GridComponent } from '../components/GridComponent';
import { COLORS } from '../constants';
import { PickComponent } from '../components/PickComponent';
import { useDiscordSDK } from '../discordsdk';

const MS_PER_MINUTE = 60_000;
const MS_PER_SECOND = 1_000;

const GameDrawTime = 1.5 * MS_PER_MINUTE;
const GameWaitTime = 1.5 * MS_PER_MINUTE;
const GameTotalTime = GameDrawTime + GameWaitTime;

export class LiveDrawScene extends Scene {
    constructor(engine) {
        super(engine)

        this.gameId = 1;
        this.nextGameTime = Date.now();
        this.picks = [];

        // Bind the websocket message callback to preserve context.
        this.onWebsocketMessage = this.onWebsocketMessage.bind(this);
    }

    resetGameState() {
        this.heads = 0;
        this.tails = 0;
        this.picks = [];
    }

    /**
    * When the scene becomes active, start a timer to update gameId every 10 seconds.
    */
    onEnter() {
        this.resetGameState();

        // Initialise the components
        this.timerCounter = new LargeCounterComponent(452, 20, "NEXT GAME", COLORS.BLUE);
        this.timerCounter.setCount(this.getTimeLeftString());

        this.drawCounter = new LargeCounterComponent(588, 20, "DRAWING GAME", COLORS.BLUE);
        this.drawCounter.setCount(this.gameId);

        this.smallCounterHeads = new SmallCounterComponent(316, 20, "Heads", COLORS.RED);
        this.smallCounterHeads.setCount(this.heads);

        this.smallCounterTails = new SmallCounterComponent(384, 20, "Tails", COLORS.BLUE);
        this.smallCounterTails.setCount(this.tails);

        this.gridComp = new GridComponent(0, 76);

        // Register the components with the scene.
        this.registerComponent(
            this.timerCounter,
            this.drawCounter,
            this.smallCounterHeads,
            this.smallCounterTails,
            this.gridComp
        );

        setInterval(() => {
            this.timerCounter.setCount(this.getTimeLeftString());
        }, 1 * MS_PER_SECOND);
    }

    /**
     * Restarts the game by removing all PickComponents,
     * resetting the pick-related state, updating counters,
     * and restarting the pick interval.
     */
    restartGame() {
        this.drawCounter.setCount(this.gameId);

        // Remove all pick components from the scene.
        this.components = this.components.filter(
            comp => !(comp instanceof PickComponent)
        );

        // Reset game state.
        this.resetGameState();

        // Reset heads and tails counters.
        this.smallCounterHeads.setCount(this.heads);
        this.smallCounterTails.setCount(this.tails);
        this.timerCounter.setCount(this.getTimeLeftString());
    }


    /**
     * Called when the scene is switched out.
     */
    onExit() {
        clearInterval(this.pickIntervalId);
        clearInterval(this.timerUpdateId);
    }

    /**
     * Returns a string representing the time left until the next game.
     * Format: MM:SS
     */
    getTimeLeftString() {
        const now = Date.now();
        const timeLeft = this.nextGameTime - now;
        if (timeLeft < 0) {
            return "00:00";
        }

        const minutes = Math.floor(timeLeft / MS_PER_MINUTE);
        const seconds = Math.floor((timeLeft - minutes * MS_PER_MINUTE) / MS_PER_SECOND);

        return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }

    onWebsocketMessage(msg) {
        if (msg.game_info) {
            this.gameId = msg.game_info.game_id;
            this.nextGameTime = Date.parse(msg.game_info.next_game_time);
            const gameStart = this.nextGameTime - GameTotalTime;

            this.restartGame();

            msg.game_info.picks?.forEach((pick) => {

                this.picks.push(pick);
                const pickComponent = new PickComponent(pick, false);
                this.registerComponent(pickComponent);

                if (pick > 40) {
                    this.tails++;
                    this.smallCounterTails.setCount(this.tails);
                } else {
                    this.heads++;
                    this.smallCounterHeads.setCount(this.heads);
                }
            });

            useDiscordSDK((sdk) =>
                sdk.commands.setActivity({
                    activity: {
                        type: 3, // "Watching {name}"
                        details: `Game ${this.gameId}`,
                        timestamps: {
                            start: gameStart,
                            end: this.nextGameTime
                        }
                    }
                }).then(() => console.log("discordsdk: set activity rich precense"))
            );

        } else if (msg.next_pick) {
            const pick = msg.next_pick.pick_number;

            this.picks.push(pick);
            const pickComponent = new PickComponent(pick);
            this.registerComponent(pickComponent);

            if (pick > 40) {
                this.tails++;
                this.smallCounterTails.setCount(this.tails);
            } else {
                this.heads++;
                this.smallCounterHeads.setCount(this.heads);
            }

        } else {
            console.warn('Received unknown message type.');
        }
    }
}

