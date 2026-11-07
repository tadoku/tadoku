import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { ChevronLeftIcon } from "../../../icons";
import { sx, type StyleXProp } from "../../../foundations/props";
import { icon, link, text } from "../../../foundations/recipes";
import { color, density, fontSize, lineHeight, measure, space, type } from "../../../tokens/tokens.stylex";
import type { NavigationLinkProps, NavigationLinkRenderer } from "../../navigation";

export interface PageHeaderProps {
  /** The page's only level-one heading. */
  readonly title: ReactNode;
  readonly lead?: ReactNode;
  /** Status tags shown beside the title, outside its accessible name. */
  readonly status?: ReactNode;
  readonly actions?: ReactNode;
  readonly backLink?: { readonly href: string; readonly label: string };
  /** Renders the back link with the application's router; defaults to a native link. */
  readonly renderLink?: NavigationLinkRenderer;
  /** Supporting content after the lead, such as a step indicator. */
  readonly children?: ReactNode;
  readonly className?: string;
  readonly xstyle?: StyleXProp;
}

const trim = `calc((${type.pageLineHeight} - ${density.controlHeight}) / 2)`;

const styles = stylex.create({
  root: { marginBlockEnd: space[32], marginBlockStart: 0, minInlineSize: 0 },
  back: {
    gap: space[4],
    textDecoration: "underline",
    alignItems: "center",
    display: "inline-flex",
    fontSize: fontSize[14],
    lineHeight: lineHeight.normal,
    marginBlockEnd: space[16],
  },
  backIcon: { transform: { default: null, ":dir(rtl)": "scaleX(-1)" } },
  main: {
    alignItems: "baseline",
    columnGap: space[24],
    display: "flex",
    flexWrap: "wrap",
    justifyContent: "space-between",
    rowGap: space[16],
  },
  text: { gap: space[8],
 display: "flex",
 flexBasis: 'auto',
 flexDirection: "column",
 flexGrow: '1',
 flexShrink: '1',
 minInlineSize: 0, },
  heading: {
    marginBlock: trim,
    alignItems: "center",
    columnGap: space[16],
    display: "flex",
    flexWrap: "wrap",
    rowGap: space[8],
  },
  status: { gap: space[8], marginBlock: trim, alignItems: "center", display: "flex", flexWrap: "wrap", },
  /** A script may focus the heading to announce the page; that focus draws no control ring. */
  title: { margin: 0, minInlineSize: 0, outlineStyle: { default: null, ":focus-visible": "none" }, overflowWrap: "anywhere" },
  lead: {
    margin: 0,
    color: color.textMuted,
    fontSize: fontSize[16],
    lineHeight: lineHeight.loose,
    maxInlineSize: measure.body,
  },
  /*
   * Actions sit top-aligned beside the title and are trimmed by the controls' overhang of its first line on each side,
   * so a single row of controls centres on that line and takes no more room than the heading row does, while wrapped
   * rows still add their height.
   */
  actions: {
    gap: density.inlineGap,
    marginBlock: `calc(${type.pageLineHeight} - ${density.controlHeight})`,
    alignItems: "center",
    alignSelf: "flex-start",
    display: "flex",
    flexWrap: "wrap",
  },
});

function nativeLink(props: NavigationLinkProps) {
  return <a {...props} />;
}

export function PageHeader({
  title,
  lead,
  status,
  actions,
  backLink,
  renderLink = nativeLink,
  children,
  className,
  xstyle,
}: PageHeaderProps) {
  return (
    <header {...sx(className, styles.root, xstyle)}>
      {backLink
        ? renderLink({
            href: backLink.href,
            ...stylex.props(link.base, styles.back),
            children: (
              <>
                <ChevronLeftIcon {...stylex.props(icon.compact, styles.backIcon)} aria-hidden="true" />
                <span>{backLink.label}</span>
              </>
            ),
          })
        : null}
      <div {...stylex.props(styles.main)}>
        <div {...stylex.props(styles.text)}>
          <div {...stylex.props(styles.heading)}>
            <h1 {...stylex.props(text.page, styles.title)}>{title}</h1>
            {status ? <div {...stylex.props(styles.status)}>{status}</div> : null}
          </div>
          {lead ? <p {...stylex.props(styles.lead)} data-page-lead="">{lead}</p> : null}
          {children}
        </div>
        {actions ? <div {...stylex.props(styles.actions)} data-page-actions="">{actions}</div> : null}
      </div>
    </header>
  );
}
