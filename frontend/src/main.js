import { Engine } from "./engine";

import { LargeCounterComponent } from './components/LargeCounterComponent.js';
import { SmallCounterComponent } from './components/SmallCounterComponent.js';
import { GridComponent } from './components/GridComponent.js';

import { DESIGN_WIDTH, DESIGN_HEIGHT, COLORS } from './constants';

/* Entry Point */
document.addEventListener('DOMContentLoaded', () => {

    /* 
     Presentation Layer: 
     - Grab the canvas element from the DOM.
     - Set its dimensions relative to the window.
    */
    const canvas = document.getElementById('canvas');
    if (!canvas) {
        console.error('Canvas element with id "canvas" not found.');
        return;
    }


    // Create the engine instance. Here our design dimensions are 720x364.
    const engine = new Engine(canvas, DESIGN_WIDTH, DESIGN_HEIGHT);

    // Create component instances:
    // A large counter component (e.g., "NEXT GAME") at position (452, 20)
    const timerCounter = new LargeCounterComponent(452, 20, "NEXT GAME", COLORS.BLUE);
    timerCounter.setCount("99:99");
   
    let gameId = 10;
    const drawCounter = new LargeCounterComponent(588, 20, "DRAWING GAME", COLORS.BLUE);
    drawCounter.setCount(gameId);

    // A small counter component for "Heads" at position (316, 20)
    const smallCounterHeads = new SmallCounterComponent(316, 20, "Heads", COLORS.RED);
    smallCounterHeads.setCount(10);

    // A small counter component for "Tails" at position (384, 20)
    const smallCounterTails = new SmallCounterComponent(384, 20, "Tails", COLORS.BLUE);
    smallCounterTails.setCount(10);

    // A grid component to display the 1-80 grid at position (0, 76)
    const gridComp = new GridComponent(0, 76);

    // Register the components with the engine.
    engine.registerComponent(timerCounter);
    engine.registerComponent(drawCounter);
    engine.registerComponent(smallCounterHeads);
    engine.registerComponent(smallCounterTails);
    engine.registerComponent(gridComp);

    // Start the engine's update and render loop.
    engine.start(); 
   
    // Update gameId by 1 every 10 seconds and update the drawCounter component.
    setInterval(() => {
        gameId += 1;
        drawCounter.setCount(gameId);
    }, 10000);
});
