// Design dimensions (in game units)
export const DESIGN_WIDTH = 720;
export const DESIGN_HEIGHT = 364;

// Color palette
export const COLORS = {
    RED: "#CA4538",
    BLUE: "#417CBF",
    GREEN: "#337B3D",
    YELLOW: "#EF8E3C",
    PINK: "#E94F83",
    ORANGE: "#EB582D",
    GRAY: "#7A8486",
    PURPLE: "#7E45C5",
    WHITE: "#FFFFFF",
    GREY_BG: "#D9D9D9",
} as const;

// Row color cycle (8 colors, indexed by row)
export const ROW_COLORS = [
    COLORS.RED,
    COLORS.BLUE,
    COLORS.GREEN,
    COLORS.YELLOW,
    COLORS.PINK,
    COLORS.ORANGE,
    COLORS.GRAY,
    COLORS.PURPLE,
] as const;

/**
 * Returns the fill color for a pick number based on its row.
 */
export function pickColor(pick: number): string {
    let row = Math.floor(pick / 10);
    if (pick % 10 === 0) {
        row--;
    }
    return ROW_COLORS[row % ROW_COLORS.length];
}
