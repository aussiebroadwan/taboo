import { Component } from "./Component";

export class RenderComponent extends Component {

    /**
     * Derived classes should implment their rendering code here.
     * @param {CanvasRenderingContext2D} ctx - Context to render on
     * @param {number} sf - The Scale Factor to help with Responsiveness
     */
    render(ctx, sf) { 
       // console.log("Renderable NOP", { ctx, sf }); 
    }
}
