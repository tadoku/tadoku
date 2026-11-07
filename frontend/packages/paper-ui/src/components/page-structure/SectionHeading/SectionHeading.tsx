import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { sx, type StyleXProp } from "../../../foundations/props";
import { text } from "../../../foundations/recipes";
import { color, density, fontSize, lineHeight, measure, space } from "../../../tokens/tokens.stylex";

export interface SectionHeadingProps {
  /** h2 and h3+ use the section and component type roles. PageHeader owns h1. */
  readonly level: 2 | 3 | 4 | 5 | 6;
  readonly title: ReactNode;
  /** Applied to the heading so the enclosing section can use aria-labelledby. */
  readonly id?: string;
  readonly description?: ReactNode;
  /** Status tags shown beside the title, outside its accessible name. */
  readonly status?: ReactNode;
  readonly actions?: ReactNode;
  readonly className?: string;
  readonly xstyle?: StyleXProp;
  /** Overrides the heading's type role, for a heading sized to its context. */
  readonly titleXstyle?: StyleXProp;
}

const styles = stylex.create({
  root: {
    alignItems: "baseline",
    columnGap: space[24],
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
    marginBlockEnd: space[16],
    marginBlockStart: 0,
    minInlineSize: 0,
    rowGap: space[12],
  },
  text: { gap: space[8],
 display: "flex",
 flexBasis: 'auto',
 flexDirection: "column",
 flexGrow: '1',
 flexShrink: '1',
 minInlineSize: 0, },
  heading: { alignItems: "center", columnGap: space[12], display: "flex", flexWrap: "wrap", rowGap: space[8] },
  cluster: { gap: density.inlineGap, alignItems: "center", display: "flex", flexWrap: "wrap", },
  /** A script may focus the heading to announce the page; that focus draws no control ring. */
  title: { margin: 0, minInlineSize: 0, outlineStyle: { default: null, ":focus-visible": "none" }, overflowWrap: "anywhere" },
  description: {
    margin: 0,
    color: color.textMuted,
    fontSize: fontSize[14],
    lineHeight: lineHeight.relaxed,
    maxInlineSize: measure.body,
  },
});

export function SectionHeading({
  level,
  title,
  id,
  description,
  status,
  actions,
  className,
  xstyle,
  titleXstyle,
}: SectionHeadingProps) {
  const Heading = `h${level}` as const;
  const typeRole = level === 2 ? text.section : text.component;
  return (
    <div {...sx(className, styles.root, xstyle)}>
      <div {...stylex.props(styles.text)}>
        <div {...stylex.props(styles.heading)}>
          <Heading id={id} {...stylex.props(typeRole, styles.title, titleXstyle)}>{title}</Heading>
          {status ? <div {...stylex.props(styles.cluster)}>{status}</div> : null}
        </div>
        {description ? <p {...stylex.props(styles.description)}>{description}</p> : null}
      </div>
      {actions ? <div {...stylex.props(styles.cluster)}>{actions}</div> : null}
    </div>
  );
}
