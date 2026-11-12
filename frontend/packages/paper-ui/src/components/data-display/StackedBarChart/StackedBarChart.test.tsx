import { render, screen } from "@testing-library/react";
import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { StackedBarChart, type StackedBarChartProps } from "./StackedBarChart";

const series = [
  { id: "reading", label: "Reading", color: "var(--color-activity-reading)" },
  { id: "listening", label: "Listening", color: "var(--color-activity-listening)" },
  { id: "speaking", label: "Speaking", color: "var(--color-activity-speaking)" },
];

const props: StackedBarChartProps = {
  start: "2026-09-01",
  end: "2026-09-10",
  series,
  days: [
    { date: "2026-09-02", values: { reading: 300, listening: 100 } },
    { date: "2026-09-09", values: { reading: 12.5 } },
    { date: "2026-10-01", values: { reading: 999 } },
  ],
  label: "Points per day by activity",
  description: "412.5 points over 10 days; the best day was 2 Sept with 400.",
};

const desktop = (container: HTMLElement) =>
  container.querySelector<SVGSVGElement>(".paper-stacked-bar-chart__plot--desktop")!;
const titles = (svg: SVGSVGElement) => Array.from(svg.querySelectorAll(".paper-stacked-bar-chart__day > title"), (title) => title.textContent);

describe("StackedBarChart", () => {
  it("is one image named and described for assistive technology", () => {
    render(<StackedBarChart {...props} />);

    const chart = screen.getByRole("img", { name: "Points per day by activity" });
    expect(chart).toHaveAccessibleDescription("412.5 points over 10 days; the best day was 2 Sept with 400.");
  });

  it("draws a slot for every calendar day in the range, including days without points, and ignores days outside it", () => {
    const { container } = render(<StackedBarChart {...props} />);

    const days = titles(desktop(container));
    expect(days).toHaveLength(10);
    expect(days[0]).toBe("1 Sept 2026: 0 points");
    expect(days.join()).not.toContain("999");
  });

  it("names the date, the total and each series' points for a day", () => {
    const { container } = render(<StackedBarChart {...props} />);

    expect(titles(desktop(container))[1]).toBe("2 Sept 2026: 400 points (Reading 300, Listening 100)");
  });

  it("stacks series in the given order with heights in proportion to their points", () => {
    const { container } = render(<StackedBarChart {...props} />);

    const [reading, listening] = Array.from(
      desktop(container).querySelectorAll(".paper-stacked-bar-chart__day")[1].querySelectorAll<SVGRectElement>("rect[data-series]"),
    );
    const top = (rect: SVGRectElement) => Number(rect.getAttribute("y"));
    const height = (rect: SVGRectElement) => Number(rect.getAttribute("height"));
    expect(reading.dataset.series).toBe("reading");
    expect(listening.dataset.series).toBe("listening");
    expect(top(listening) + height(listening)).toBeCloseTo(top(reading));
    expect(height(reading) / height(listening)).toBeCloseTo(3);
  });

  it("labels the points axis at zero, half and a round maximum at or above the best day", () => {
    const { container } = render(<StackedBarChart {...props} />);

    const labels = Array.from(desktop(container).querySelectorAll(".paper-stacked-bar-chart__axis--points"), (text) =>
      text.textContent,
    );
    expect(labels).toEqual(["0", "200", "400"]);
  });

  it("adds a running total on its own axis and in the legend only when asked", () => {
    const { container, rerender } = render(<StackedBarChart {...props} />);
    expect(screen.queryByText("Running total")).not.toBeInTheDocument();
    expect(desktop(container).querySelector(".paper-stacked-bar-chart__total")).toBeNull();

    rerender(<StackedBarChart {...props} days={[...props.days, { date: "2026-09-10", values: { reading: 11800 } }]} runningTotal />);
    const totals = Array.from(desktop(container).querySelectorAll(".paper-stacked-bar-chart__axis--total"), (text) =>
      text.textContent,
    );
    expect(totals).toEqual(["0", "7.5k", "15k"]);
    expect(desktop(container).querySelector(".paper-stacked-bar-chart__total")).not.toBeNull();
    expect(titles(desktop(container))[9]).toBe("10 Sept 2026: 11,800 points, 12,212.5 in total (Reading 11,800)");
    expect(screen.getByText("Running total")).toBeInTheDocument();
  });

  it("lists only the series with points in the legend", () => {
    render(<StackedBarChart {...props} />);

    expect(screen.getByText("Reading")).toBeInTheDocument();
    expect(screen.getByText("Listening")).toBeInTheDocument();
    expect(screen.queryByText("Speaking")).not.toBeInTheDocument();
  });

  it("labels short ranges by day and long ranges by month", () => {
    const { container, rerender } = render(<StackedBarChart {...props} />);
    const ticks = () =>
      Array.from(desktop(container).querySelectorAll(".paper-stacked-bar-chart__axis--date"), (text) => text.textContent);
    expect(ticks()).toEqual(["1 Sept", "4 Sept", "7 Sept", "10 Sept"]);

    rerender(<StackedBarChart {...props} start="2026-09-15" end="2026-12-31" />);
    expect(ticks()).toEqual(["Oct", "Nov", "Dec"]);
  });

  it("draws an empty range without failing", () => {
    const { container } = render(<StackedBarChart {...props} days={[]} runningTotal />);

    expect(titles(desktop(container))).toHaveLength(10);
    expect(container.innerHTML).not.toContain("NaN");
  });

  it("renders on the server without a DOM", () => {
    const html = renderToString(<StackedBarChart {...props} runningTotal />);

    expect(html).toContain('role="img"');
    expect(html).toContain("2 Sept 2026: 400 points");
  });
});
