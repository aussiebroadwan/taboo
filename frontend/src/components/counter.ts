/**
 * Large counter element (e.g. "NEXT GAME", "DRAWING GAME").
 */
export interface Counter {
    element: HTMLElement;
    setValue(value: string | number): void;
}

export function createLargeCounter(title: string, color: string): Counter {
    const el = document.createElement("div");
    el.className = "counter-large";
    el.style.backgroundColor = color;

    const titleEl = document.createElement("span");
    titleEl.className = "counter-large__title";
    titleEl.textContent = title;

    const valueEl = document.createElement("span");
    valueEl.className = "counter-large__value";
    valueEl.textContent = "0";

    el.appendChild(titleEl);
    el.appendChild(valueEl);

    return {
        element: el,
        setValue(value: string | number) {
            valueEl.textContent = String(value);
        },
    };
}

/**
 * Small counter element (e.g. "Heads", "Tails").
 */
export function createSmallCounter(title: string, color: string): Counter {
    const el = document.createElement("div");
    el.className = "counter-small";

    const titleEl = document.createElement("span");
    titleEl.className = "counter-small__title";
    titleEl.style.color = color;
    titleEl.textContent = title;

    const box = document.createElement("div");
    box.className = "counter-small__box";
    box.style.backgroundColor = color;

    const valueEl = document.createElement("span");
    valueEl.className = "counter-small__value";
    valueEl.textContent = "0";

    box.appendChild(valueEl);
    el.appendChild(titleEl);
    el.appendChild(box);

    return {
        element: el,
        setValue(value: string | number) {
            valueEl.textContent = String(value);
        },
    };
}
