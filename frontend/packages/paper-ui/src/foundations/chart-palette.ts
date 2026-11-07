import { color } from "../tokens/tokens.stylex";

/**
 * Chart color roles resolve through the active Paper theme. The pattern and
 * dash metadata keep adjacent categories distinguishable without hue alone.
 */
export const chartPalette = [
  color.chart1,
  color.chart2,
  color.chart3,
  color.chart4,
  color.chart5,
  color.chart6,
  color.chart7,
  color.chart8,
] as const;

export const chartCategoryStyles = chartPalette.map((value, index) => ({
  color: value,
  dash: [[], [6, 3], [2, 2], [8, 3, 2, 3]][index % 4] as readonly number[],
  pattern: ["solid", "diagonal", "dots", "crosshatch"][index % 4] as
    | "solid"
    | "diagonal"
    | "dots"
    | "crosshatch",
  pointStyle: ["circle", "rect", "triangle", "rectRot"][index % 4] as
    | "circle"
    | "rect"
    | "triangle"
    | "rectRot",
}));

export type ChartCategoryStyle = (typeof chartCategoryStyles)[number];

/** Backward-readable name for consumers that describe chart categories as series. */
export const chartSeries = chartCategoryStyles;
export type ChartSeries = ChartCategoryStyle;
