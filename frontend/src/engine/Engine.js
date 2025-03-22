

export class Engine {
    /**
    * @param {HTMLCanvasElement} canvas - The canvas element to render on.
    * @param {number} designWidth - The design width in game units (default: 720).
    * @param {number} designHeight - The design height in game units (default: 364).
    */
    constructor(canvas, designWidth = 720, designHeight = 364) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.designWidth = designWidth;
        this.designHeight = designHeight;
        this.sf = 1; // scale factor (calculated from canvas container size)
        this.components = []; // registered components
        this.lastTime = performance.now();
        this.renderFrameId = null;

        // Bind the resize handler and perform an initial resize.
        window.addEventListener('resize', this.updateCanvasSize.bind(this));
        this.updateCanvasSize();
    }

    /**
     * Updates the canvas size based on its container's width,
     * maintains the design aspect ratio, and computes the new scale factor.
     */
    updateCanvasSize() {
        const container = this.canvas.parentElement;
        const dpr = window.devicePixelRatio || 1;
        const cssWidth = container.offsetWidth;
        const cssHeight = cssWidth * (this.designHeight / this.designWidth);

        // Set the CSS display size.
        this.canvas.style.width = `${cssWidth}px`;
        this.canvas.style.height = `${cssHeight}px`;

        // Set the internal pixel resolution.
        this.canvas.width = cssWidth * dpr;
        this.canvas.height = cssHeight * dpr;

        // Reset and scale the drawing context to ensure crisp rendering.
        this.ctx.resetTransform();
        this.ctx.scale(dpr, dpr);

        // Calculate the scale factor used for all drawing operations.
        this.sf = cssWidth / this.designWidth;
    }

    /**
     * Register a render component. The component is expected to implement:
     * - update(dt)  for state updates.
     * - render(ctx, sf) for drawing.
     * @param {Object} component - The component instance.
     */
    registerComponent(component) {
        this.components.push(component);
    }

    /**
     * Update each registered component.
     * @param {number} dt - Delta time (in milliseconds) since the last update.
     */
    update(dt) {
        for (const comp of this.components) {
            if (typeof comp.update === 'function') {
                comp.update(dt);
            }
        }
    }

    /**
     * Clears the canvas and renders all registered components.
     */
    render() {
        // Clear the entire canvas.
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        for (const comp of this.components) {
             if (typeof comp.render === 'function') {
                comp.render(this.ctx, this.sf);
             }
        }
    }

    /**
     * Main loop callback.
     * @param {DOMHighResTimeStamp} timestamp - The current timestamp.
     */
    loop = (timestamp) => {
        const dt = timestamp - this.lastTime;
        this.lastTime = timestamp;

        // Update component states.
        this.update(dt);

        // Render components to the canvas.
        this.render();

        // Queue next frame.
        this.renderFrameId = requestAnimationFrame(this.loop);
    }

    /**
     * Starts the main update/render loop.
     */
    start() {
        this.lastTime = performance.now();
        this.renderFrameId = requestAnimationFrame(this.loop);
    }

    /**
     * Stops the main loop.
     */
    stop() {
        if (this.renderFrameId) {
            cancelAnimationFrame(this.renderFrameId);
            this.renderFrameId = null;
        }
    }
}
