import { pickColor } from "../constants";

/**
 * Instantly place a pick on the grid (no animation).
 * Used for initial state load and previous draw view.
 */
export function placePickInstant(cells: Map<number, HTMLElement>, pick: number): void {
    const cell = cells.get(pick);
    if (!cell) return;
    cell.style.backgroundColor = pickColor(pick);
    cell.classList.add("grid-cell--picked");
}

/**
 * Animate a pick from the center of the container to its grid cell.
 * Uses the Web Animations API.
 */
export function placePickAnimated(
    container: HTMLElement,
    cells: Map<number, HTMLElement>,
    pick: number,
): void {
    const cell = cells.get(pick);
    if (!cell) return;

    const color = pickColor(pick);

    // Get container and cell rects for positioning
    const containerRect = container.getBoundingClientRect();
    const cellRect = cell.getBoundingClientRect();

    // Starting position: center of container
    const startSize = Math.min(containerRect.width * 0.4, 280);
    const startLeft = (containerRect.width - startSize) / 2;
    const startTop = (containerRect.height - startSize) / 2;

    // End position: cell position relative to container
    const endLeft = cellRect.left - containerRect.left;
    const endTop = cellRect.top - containerRect.top;
    const endWidth = cellRect.width;
    const endHeight = cellRect.height;

    // Create the flying element
    const fly = document.createElement("div");
    fly.className = "pick-fly";
    fly.textContent = String(pick);
    fly.style.backgroundColor = color;
    fly.style.fontSize = `${startSize * 0.45}px`;
    fly.style.width = `${startSize}px`;
    fly.style.height = `${startSize}px`;
    fly.style.borderRadius = "50%";
    fly.style.left = `${startLeft}px`;
    fly.style.top = `${startTop}px`;
    container.appendChild(fly);

    // End font size must match grid-cell CSS: font-size 2.5cqi = 2.5% of container width
    const endFontSize = containerRect.width * 0.025;

    // Hold in center for 1s, then fly to cell over 1s
    const startFrame = {
        left: `${startLeft}px`,
        top: `${startTop}px`,
        width: `${startSize}px`,
        height: `${startSize}px`,
        borderRadius: "50%",
        fontSize: `${startSize * 0.45}px`,
    };

    const animation = fly.animate(
        [
            { ...startFrame, offset: 0 },
            { ...startFrame, offset: 0.5 },
            {
                left: `${endLeft}px`,
                top: `${endTop}px`,
                width: `${endWidth}px`,
                height: `${endHeight}px`,
                borderRadius: "0.28cqi",
                fontSize: `${endFontSize}px`,
                offset: 1,
            },
        ],
        {
            duration: 2000,
            easing: "cubic-bezier(0.455, 0.03, 0.515, 0.955)",
            fill: "forwards",
        },
    );

    animation.onfinish = () => {
        fly.remove();
        cell.style.backgroundColor = color;
        cell.classList.add("grid-cell--picked");
    };
}
