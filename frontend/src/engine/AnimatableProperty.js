
export class AnimatableProperty {

    constructor() {
        this.keyframes = [];
        this.easing = Ease.linear;
        this.duration = 0;
        this.elapsed = 0;
    }

    /**
     * Adds a keyframe at a given time with a specific value.
     * @param {number} time - Time in milliseconds.
     * @param {number} value - Value at that time.
     * @returns {AnimatableProperty} Returns this instance for chaining.
     */
    addStep(time, value) {
        this.keyframes.push({ time, value });
        // Keep keyframes sorted by time.
        this.keyframes.sort((a, b) => a.time - b.time);
        // Update the overall duration.
        this.duration = Math.max(this.duration, time);
        return this;
    }

    /**
     * Sets the easing function for interpolation.
     * @param {function} easeFn - A function that takes a normalized time t (0 to 1) and returns an eased value.
     * @returns {AnimatableProperty} Returns this instance for chaining.
     */
    setEase(easeFn) {
        this.easing = easeFn;
        return this;
    }


    /**
     * Update the property by advancing time.
     * @param {number} dt - Delta time in milliseconds.
     */
    update(dt) {
        this.elapsed = Math.min(this.elapsed + dt, this.duration);
    }

    /**
     * Reset the property animation.
     */
    reset() {
        this.elapsed = 0;
    }

    /**
     * Get the current interpolated value.
     * @returns {number}
     */
    get() {
        // If before the first keyframe, return the first value.
        if (this.elapsed <= this.keyframes[0].time) {
            return this.keyframes[0].value;
        }

        // Find the keyframes that surround the current elapsed time.
        for (let i = 1; i < this.keyframes.length; i++) {
            const prev = this.keyframes[i - 1];
            const next = this.keyframes[i];
            if (this.elapsed <= next.time) {
                const t = (this.elapsed - prev.time) / (next.time - prev.time);
                const easedT = this.easing(t);
                return prev.value + (next.value - prev.value) * easedT;
            }
        }

        // If elapsed time exceeds duration, return final value.
        return this.keyframes[this.keyframes.length - 1].value;
    }
}

export const Ease = {
  /**
   * Linear easing (no acceleration)
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  linear(t) {
    return t;
  },

  /**
   * Ease In Quad: accelerating from zero velocity.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  InQuad(t) {
    return t * t;
  },

  /**
   * Ease Out Quad: decelerating to zero velocity.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  OutQuad(t) {
    return t * (2 - t);
  },

  /**
   * Ease In Out Quad: acceleration until halfway, then deceleration.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  InOutQuad(t) {
    return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t;
  },

  /**
   * Ease In Cubic: accelerating from zero velocity.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  InCubic(t) {
    return t * t * t;
  },

  /**
   * Ease Out Cubic: decelerating to zero velocity.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  OutCubic(t) {
    return (--t) * t * t + 1;
  },

  /**
   * Ease In Out Cubic: acceleration until halfway, then deceleration.
   * @param {number} t - Normalized time (0 to 1)
   * @returns {number}
   */
  InOutCubic(t) {
    return t < 0.5 ? 4 * t * t * t : (t - 1) * (2 * t - 2) * (2 * t - 2) + 1;
  }
};
