import { RenderComponent, AnimatableProperty, Ease } from "../engine";
import { DESIGN_WIDTH, DESIGN_HEIGHT, COLORS } from "../constants";


const fillColor = (pick) => {
    const row = Math.floor(pick / 10);
    const colors = [
        COLORS.RED,
        COLORS.BLUE,
        COLORS.GREEN,
        COLORS.YELLOW,
        COLORS.PINK,
        COLORS.ORANGE,
        COLORS.GRAY,
        COLORS.PURPLE
    ];

    if (pick % 10 == 0) {
        return colors[(row - 1) % colors.length];
    }

    return colors[row % colors.length];
}

const endPosition = (pick) => {
    const sx = 76;
    const sy = 92;

    let row = Math.floor(pick / 10);
    let col = pick % 10 - 1;

    if (col < 0) {
        col = 9;
        row--;
    }

    const yOff = pick > 40 ? 4 : 0;

    return {
        x: sx + (col * (64 + 4)),
        y: sy + (row * (32 + 4)) + yOff
    }
}

export class PickComponent extends RenderComponent {

    /**
     * @param {number|string} pickNumber - The number to display.
     * @param {number} duration - Duration of the entire animation (including delay) in milliseconds.
     * @param {number} delay - Delay (in ms) to wait at the start before beginning the transition.
     */
    constructor(pickNumber, animate = true, duration = 1500, delay = 1000) {
        super();

        this.pickNumber = pickNumber;

        this.startX = DESIGN_WIDTH / 2;
        this.startY = DESIGN_HEIGHT / 2;

        const endPos = endPosition(pickNumber);
        this.endX = endPos.x;
        this.endY = endPos.y;

        this.duration = duration;
        this.delay = delay;
        this.elapsed = 0;
        this.animationComplete = false;

        this.fillcolor = fillColor(pickNumber);

        if (animate) {
            this.positionX = new AnimatableProperty()
                .addStep(0, DESIGN_WIDTH / 2)
                .addStep(duration, endPos.x)
                .setEase(Ease.InOutQuad);

            this.positionY = new AnimatableProperty()
                .addStep(0, DESIGN_HEIGHT / 2)
                .addStep(duration, endPos.y)
                .setEase(Ease.InOutQuad);

            this.radius = new AnimatableProperty()
                .addStep(0, 100)
                .addStep(duration - 150, 10)
                .addStep(duration, 64)
                .setEase(Ease.InOutQuad);

            this.fontSize = new AnimatableProperty()
                .addStep(0, 64)
                .addStep(duration - 150, 10)
                .addStep(duration, 18)
                .setEase(Ease.InOutQuad);

            this.clipWidth = new AnimatableProperty()
                .addStep(0, 200)
                .addStep(duration, 64)
                .setEase(Ease.InOutQuad);

            this.clipHeight = new AnimatableProperty()
                .addStep(0, 200)
                .addStep(duration, 32)
                .setEase(Ease.InOutQuad);
        } else {
            this.positionX = new AnimatableProperty()
                .addStep(duration, endPos.x);

            this.positionY = new AnimatableProperty()
                .addStep(duration, endPos.y);

            this.radius = new AnimatableProperty()
                .addStep(duration, 64);

            this.fontSize = new AnimatableProperty()
                .addStep(duration, 18);

            this.clipWidth = new AnimatableProperty()
                .addStep(duration, 64);

            this.clipHeight = new AnimatableProperty()
                .addStep(duration, 32);
        }
    }

    /**
     * Update the animation based on delta time.
     * @param {number} dt - Delta time in milliseconds.
     */
    update(dt) {
        if (this.animationComplete) return;
        this.elapsed += dt;
        if (this.elapsed >= this.duration + this.delay) {
            this.elapsed = this.duration + this.delay;
            this.animationComplete = true;
        }

        // Only update the animatable properties after the delay.
        if (this.elapsed > this.delay) {
            this.positionX.update(dt);
            this.positionY.update(dt);
            this.radius.update(dt);
            this.fontSize.update(dt);
            this.clipHeight.update(dt);
            this.clipWidth.update(dt);
        }
    }

    /**
     * Render the animated pick.
     * @param {CanvasRenderingContext2D} ctx - The canvas context.
     * @param {number} sf - The scale factor.
     */
    render(ctx, sf) {
        const currentX = this.positionX.get();
        const currentY = this.positionY.get();
        const currentRadius = this.radius.get();
        const currentFontSize = this.fontSize.get();
        const clipWidth = this.clipWidth.get();
        const clipHeight = this.clipHeight.get();

        ctx.save();
        ctx.translate(currentX * sf, currentY * sf);

        // Set up the clipping region (centered at (0,0)).
        ctx.beginPath();
        ctx.roundRect(
            -clipWidth / 2 * sf,
            -clipHeight / 2 * sf,
            clipWidth * sf,
            clipHeight * sf,
            [2 * sf]
        );
        ctx.clip();

        // Draw the circle inside the clipped region.
        ctx.beginPath();
        ctx.arc(0, 0, currentRadius * sf, 0, Math.PI * 2);
        ctx.fillStyle = this.fillcolor;
        ctx.fill();
        ctx.closePath();

        // Draw the pick number text on top (centered)
        ctx.fillStyle = COLORS.WHITE;
        ctx.font = `bold ${currentFontSize * sf}px Arial`;
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(this.pickNumber, 0, 0);
        ctx.restore();
    }
}
