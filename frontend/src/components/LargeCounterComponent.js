import { RenderComponent } from "../engine";
import { COLORS } from "../constants";

/**
 * An instance of the Large Counter Block
 */
export class LargeCounterComponent extends RenderComponent {

    /**
     * Create an instance of the Large Counter block (Next Game/Game Number). 
     * This is sized as 132x52 units.
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
         * @type {string|number}
         */
        this._count = 0; 
    }

    /**
     * Set the new count to display.
     * @param {string|number} new_count - The new count value to display.
     * @return {void}
     */
    setCount = ( new_count ) => this._count = new_count;

    render(ctx, sf) {
        ctx.save();
        ctx.translate(this.x * sf, this.y * sf);
            ctx.beginPath();
                ctx.fillStyle = this.color;
                ctx.roundRect(0, 0, 132 * sf, 52 * sf, [2 * sf]);
                ctx.fill();

                // Count text
                ctx.font = 'bold ' + (24 * sf) + 'px Arial';
                ctx.fillStyle = COLORS.WHITE;
                ctx.textAlign = 'center';
                ctx.textBaseline = 'middle';
                ctx.fillText(`${this._count}`, 66 * sf, 37 * sf);

                // Header text
                ctx.font = 'bold ' + (10 * sf) + 'px Arial';
                ctx.fillStyle = COLORS.WHITE;
                ctx.fillText(this.title, 66 * sf, 14 * sf);
            ctx.closePath();
        ctx.restore();
    }
}

