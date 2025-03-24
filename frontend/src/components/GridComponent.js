import { RenderComponent } from "../engine";
import { COLORS } from "../constants";

/**
 * An instance of the 1-80 Grid
 */
export class GridComponent extends RenderComponent {

    /**
     * Create an instance of the Grid that displays the 1-80 cells. 
     * This is of size 720x288 units.
     *
     * @param {number} x - The Top Left 'x' Position of the block
     * @param {number} y - The Top Left 'y' Position of the block
     */
    constructor(x, y) {
        super();

        this.x = x;
        this.y = y;
    }

    render(ctx, sf) {

        // Pre-Render the Set of 40 cells for each half
        const grid = new Path2D();
        for (let x = 0; x < 10; x++) {
            for (let y = 0; y < 4; y++) {
                // Spacing is 64px square plus 4px gap in design coordinates.
                const posX = x * (64 + 4) * sf;
                const posY = y * (32 + 4) * sf;
                grid.roundRect(posX, posY, 64 * sf, 32 * sf, [2 * sf]);
            }
        }
        grid.closePath();


        // Render HEADS Top Grid 1-40 (720x140)
        ctx.save();
        ctx.translate(this.x, this.y * sf);
            ctx.beginPath();
                // Left block background
                ctx.fillStyle = COLORS.RED;
                ctx.roundRect(0, 0, 40 * sf, 140 * sf, [2 * sf]);
                ctx.fill();

                // Rotated header ("Heads")
                ctx.save();
                    ctx.translate(20 * sf, 70 * sf);
                    ctx.rotate(-Math.PI / 2);
                    ctx.font = 'bold ' + (18 * sf) + 'px Arial';
                    ctx.fillStyle = COLORS.WHITE;
                    ctx.textAlign = 'center';
                    ctx.textBaseline = 'middle';
                    ctx.fillText("Heads", 0, 0);
                ctx.restore();

                // Render grid of rounded rectangles.
                ctx.fillStyle = COLORS.GREY_BG;
                ctx.translate(44 * sf, 0);
                ctx.fill(grid);
            ctx.closePath();
        ctx.restore();

        // Render TAILS Bottom Grid 41-80 (720x140)
        ctx.save();
        ctx.translate(this.x, (this.y + 8 + 140)* sf);
            ctx.beginPath();
                // Left block background
                ctx.fillStyle = COLORS.BLUE;
                ctx.roundRect(0, 0, 40 * sf, 140 * sf, [2 * sf]);
                ctx.fill();

                // Rotated header ("Tails")
                ctx.save();
                    ctx.translate(20 * sf, 70 * sf);
                    ctx.rotate(-Math.PI / 2);
                    ctx.font = 'bold ' + (18 * sf) + 'px Arial';
                    ctx.fillStyle = COLORS.WHITE;
                    ctx.textAlign = 'center';
                    ctx.textBaseline = 'middle';
                    ctx.fillText("Tails", 0, 0);
                ctx.restore();

                // Render grid of rounded rectangles.
                ctx.fillStyle = COLORS.GREY_BG;
                ctx.translate(44 * sf, 0);
                ctx.fill(grid);
            ctx.closePath();
        ctx.restore();
    }
}

