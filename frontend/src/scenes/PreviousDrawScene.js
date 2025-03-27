import { Scene } from '../engine';
import { LargeCounterComponent } from '../components/LargeCounterComponent';
import { SmallCounterComponent } from '../components/SmallCounterComponent';
import { GridComponent } from '../components/GridComponent';
import { COLORS } from '../constants';
import { PickComponent } from '../components/PickComponent';

import { usingDiscordSDK } from '../discordsdk';

const protocol = window.location.protocol;
const hostname = window.location.host;

export class PreviousDrawScene extends Scene {
    constructor(engine, gameId) {
        super(engine)

        this.gameId = gameId;
        this.picks = [];
        this.heads = 0;
        this.tails = 0;
    }

    onEnter() {
        // Initialise the components
        this.timerCounter = new LargeCounterComponent(452, 20, "NEXT GAME", COLORS.BLUE);
        this.timerCounter.setCount("00:00");

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

        const apiCall = `${protocol}//${hostname}/${usingDiscordSDK ? ".proxy/" : ""}/api/game/${this.gameId}`;

        fetch(apiCall)
            .then(response => {
                if (response.status != 200) {
                    response.text()
                        .then(data => {
                            console.log('previous_draw_scene: unable to get draw redirecting to live', data)
                            window.location.href = '/';
                        })
                } else {
                    response.json()
                        .then(data => {
                            data.picks.forEach((pick) => {

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
                            console.log(`previous_draw_scene: drawn previous game ${this.gameId}`);
                        })
                        .catch(error => {
                            console.log('previous_draw_scene: unable to parse api response', error)
                            window.location.href = '/';
                        });
                }
            })
            .catch(error => {
                console.log('previous_draw_scene: unable to get response redirecting to live', error)
                window.location.href = '/';
            });

    }
}

