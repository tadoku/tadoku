import * as stylex from "@stylexjs/stylex";
import { blur, breakpoints, color, motion, opacity, zIndex } from "../../tokens/tokens.stylex";


/** The dimmed layer behind Modal and Drawer. */
export const backdrop = stylex.create({
  root: {
    inset: 0,
    backgroundColor: { default: color.surfaceScrim, [breakpoints.forcedColors]: color.forcedScrim },
    forcedColorAdjust: { default: null, [breakpoints.forcedColors]: "none" },
    opacity: { "[data-ending-style]": 0, "[data-starting-style]": 0, default: opacity.visible, },
    position: "fixed",
    transitionDuration: motion.standard,
    transitionProperty: { default: "opacity", [breakpoints.reducedMotion]: "none" },
    zIndex: zIndex.scrim,
  },
  blurred: {
    backdropFilter: { default: `blur(${blur.scrim})`, [breakpoints.forcedColors]: "none" },
  },
});
