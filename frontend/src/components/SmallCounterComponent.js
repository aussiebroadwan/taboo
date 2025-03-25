import { RenderComponent } from "../engine";
import { COLORS } from "../constants";

/**
 * An instance of the Small Counter Block
 */
export class SmallCounterComponent extends RenderComponent {

    /**
     * Create an instance of the Small Counter block (HEADS/TAILS). This is sized 
     * as 64x52 units.
     *
     * @param {number} x - The Top Left 'x' Position of the block
     * @param {number} y - The Top Left 'y' Position of the block
     * @param {string} title - The Text to Display as the Title
     * @param {string} color - The Color style text to set the block and title
     */
    constructor(x, y, title, color) {
        super();

        this.x = x;
        this.y = y;
        this.title = title;
        this.color = color;
       
        /**
         * The Count to display in the block 
         * @type {number|string}
         */
        this._count = 0; 
    }

    /**
     * Set the new count to display.
     * @param {number|string} new_count - The new count value to display.
     * @return {void}
     */
    setCount = ( new_count ) => this._count = new_count;

    render(ctx, sf) {
        ctx.save();
        ctx.translate(this.x * sf, this.y * sf);
            ctx.beginPath();
                // Rounded rectangle at (0, 20)
                ctx.fillStyle = this.color;
                ctx.roundRect(0, 20 * sf, 64 * sf, 32 * sf, [2 * sf]);
                ctx.fill();

                // Header text
                ctx.font = 'bold ' + (10 * sf) + 'px Arial';
                ctx.fillStyle = this.color;
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.fillText(this.title, 32 * sf, 12 * sf);

                // Count text
                ctx.font = 'bold ' + (24 * sf) + 'px Arial';
                ctx.fillStyle = COLORS.WHITE;
                ctx.fillText(`${this._count}`, 32 * sf, 37 * sf );
            ctx.closePath();
        ctx.restore();
    }
}
