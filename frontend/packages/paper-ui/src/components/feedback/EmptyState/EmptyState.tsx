import * as stylex from "@stylexjs/stylex";
import type { HTMLAttributes, ReactNode, Ref } from "react";
import { sx, type StyleXProp } from "../../../foundations/props";
import { text } from "../../../foundations/recipes";
import { border, color, density, measure, space } from "../../../tokens/tokens.stylex";

export interface EmptyStateProps extends Omit<HTMLAttributes<HTMLDivElement>, "title"> {
  readonly title: ReactNode;
  /** Where the state sits in the page outline: 1 when it replaces the page. */
  readonly headingLevel: 1 | 2 | 3 | 4 | 5 | 6;
  readonly description?: ReactNode;
  readonly actions?: ReactNode;
  /** `surface` stands alone on paper; `plain` sits inside an existing container. */
  readonly variant?: "surface" | "plain";
  /** Receives the heading, which becomes programmatically focusable for focus moves. */
  readonly headingRef?: Ref<HTMLHeadingElement>;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  root: { gap: space[12], alignItems: "flex-start", display: "flex", flexDirection: "column", minInlineSize: 0, },
  surface: {
    borderColor: color.ruleDefault,
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: space[32],
    paddingInline: space[24],
    backgroundColor: color.surfacePaper,
  },
  plain: { paddingBlock: space[24] },
  title: {
    margin: 0,
    maxInlineSize: "100%",
    outlineStyle: {
      default: null,
      ":focus": "none",
    },
    overflowWrap: "anywhere",
  },
  description: { margin: 0, maxInlineSize: measure.lead },
  actions: {
    gap: density.inlineGap,
    alignItems: "center",
    display: "flex",
    flexWrap: "wrap",
    marginBlockStart: space[8],
  },
});

export function EmptyState({
  title,
  headingLevel,
  description,
  actions,
  variant = "surface",
  headingRef,
  className,
  xstyle,
  ...props
}: EmptyStateProps) {
  const Heading = `h${headingLevel}` as const;
  const typeRole = headingLevel === 1 ? text.page : headingLevel === 2 ? text.section : text.component;
  return (
    <div {...props} {...sx(className, styles.root, styles[variant], xstyle)}>
      <Heading ref={headingRef} tabIndex={headingRef ? -1 : undefined} {...stylex.props(typeRole, styles.title)}>
        {title}
      </Heading>
      {description ? <p {...stylex.props(styles.description)}>{description}</p> : null}
      {actions ? <div {...stylex.props(styles.actions)}>{actions}</div> : null}
    </div>
  );
}
