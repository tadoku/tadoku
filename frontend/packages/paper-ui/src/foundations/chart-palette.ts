/**
 * Chart color roles resolve through the active Paper theme. The pattern and
 * dash metadata keep adjacent categories distinguishable without hue alone.
 */
export const chartPalette = [
  "var(--paper-color-chart-1)",
  "var(--paper-color-chart-2)",
  "var(--paper-color-chart-3)",
  "var(--paper-color-chart-4)",
  "var(--paper-color-chart-5)",
  "var(--paper-color-chart-6)",
  "var(--paper-color-chart-7)",
  "var(--paper-color-chart-8)",
] as const;

export const chartCategoryStyles = chartPalette.map((color, index) => ({
  color,
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
