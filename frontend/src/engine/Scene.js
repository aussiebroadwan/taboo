import { Component } from "./Component";
import { RenderComponent } from "./RenderComponent";

export class Scene {
    /**
     * @param {Engine} engine - A reference to the game engine.
     */
    constructor(engine) {
        this.engine = engine;
        this.components = [];
    }

    /**
     * Register a component to be updated and rendered by the scene.
     * @param {...Object} component - The component must implement update(dt) and render(ctx, sf).
     */
    registerComponent(...component) {
        this.components.push(...component);
    }

    /**
     * Remove a component from the scene.
     * @param {Object} component - The component instance to remove.
     */
    removeComponent(component) {
        const index = this.components.indexOf(component);
        if (index !== -1) {
            this.components.splice(index, 1);
        }
    }

    /**
     * Clears all components from the scene.
     */
    clearComponents() {
        this.components.length = 0;
    }

    /**
     * Update all components in the scene.
     * @param {number} dt - Delta time in milliseconds.
     */
    update(dt) {
        this.components
            .filter(comp => comp instanceof Component)
            .forEach(comp => comp.update(dt));
    }

    /**
     * Render all components in the scene.
     * @param {CanvasRenderingContext2D} ctx - The drawing context.
     * @param {number} sf - The scale factor.
     */
    render(ctx, sf) {
        this.components
            .filter(comp => comp instanceof RenderComponent)
            .forEach(comp => comp.render(ctx, sf));
    }

    /**
     * Called when the scene becomes active.
     */
    onEnter() {
        // Can be overridden in subclasses.
    }

    /**
     * Called when the scene is switched away.
     */
    onExit() {
        // Can be overridden in subclasses.
    }
}
