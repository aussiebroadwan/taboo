import { COLORS } from "../constants";

/**
 * Builds two 10x4 grids (Heads 1-40, Tails 41-80) and returns:
 * - The root element to append to the container
 * - A Map from pick number (1-80) to its cell HTMLElement
 */
export function createGrid(): { element: HTMLElement; cells: Map<number, HTMLElement> } {
    const cells = new Map<number, HTMLElement>();

    const section = document.createElement("div");
    section.className = "grid-section";

    section.appendChild(buildHalf("Heads", COLORS.RED, 1, cells));
    section.appendChild(buildHalf("Tails", COLORS.BLUE, 41, cells));

    return { element: section, cells };
}

function buildHalf(
    label: string,
    labelColor: string,
    startNum: number,
    cells: Map<number, HTMLElement>,
): HTMLElement {
    const half = document.createElement("div");
    half.className = "grid-half";

    // Vertical label
    const labelEl = document.createElement("div");
    labelEl.className = "grid-label";
    labelEl.style.backgroundColor = labelColor;
    const labelText = document.createElement("span");
    labelText.className = "grid-label__text";
    labelText.textContent = label;
    labelEl.appendChild(labelText);
    half.appendChild(labelEl);

    // 10x4 grid of cells
    const gridEl = document.createElement("div");
    gridEl.className = "grid-cells";

    for (let row = 0; row < 4; row++) {
        for (let col = 0; col < 10; col++) {
            const num = startNum + row * 10 + col;
            const cell = document.createElement("div");
            cell.className = "grid-cell";
            cell.textContent = String(num);
            cell.dataset.num = String(num);
            gridEl.appendChild(cell);
            cells.set(num, cell);
        }
    }

    half.appendChild(gridEl);
    return half;
}
