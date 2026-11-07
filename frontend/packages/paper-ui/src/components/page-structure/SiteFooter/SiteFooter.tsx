import * as stylex from "@stylexjs/stylex";
import type { ReactNode } from "react";
import { fillIcon } from "../../../foundations/IconSlot";
import type { StyleXProp } from "../../../foundations/props";
import {
  border,
  breakpoints,
  color,
  decoration,
  font,
  fontSize,
  fontWeight,
  lineHeight,
  page,
  size,
  space,
  tracking,
} from "../../../tokens/tokens.stylex";
import { renderNavigationLink, type NavigationLinkRenderer } from "../../navigation/navigation";

export interface SiteFooterLink {
  readonly id: string;
  readonly label: string;
  readonly href: string;
}

export interface SiteFooterProps {
  /** The brand lockup, linked to `brandHref` and named by `brandLabel`. */
  readonly brand: ReactNode;
  readonly brandHref: string;
  readonly brandLabel: string;
  /** A short line under the brand, such as a credit; links in it take the band's link style. */
  readonly credit?: ReactNode;
  /** External community links shown as icon buttons. */
  readonly social?: readonly { readonly label: string; readonly href: string; readonly icon: ReactNode }[];
  readonly socialLabel?: string;
  readonly groups: readonly { readonly title: string; readonly links: readonly SiteFooterLink[] }[];
  readonly label?: string;
  /** Renders internal links with the application's router; defaults to native links. */
  readonly renderLink?: NavigationLinkRenderer;
  readonly xstyle?: StyleXProp;
}

const styles = stylex.create({
  /** An ink band in light mode; in dark mode it lifts to the paper surface instead of sinking below the canvas. */
  band: {
    paddingInline: page.gutter,
    backgroundColor: color.bandSurface,
    color: color.bandText,
    paddingBlockEnd: space[56],
    paddingBlockStart: space[48],
  },
  inner: {
    marginInline: "auto",
    columnGap: space[64],
    display: "grid",
    gridTemplateColumns: { default: "minmax(0, 1fr) auto", [breakpoints.narrow]: "minmax(0, 1fr)" },
    maxInlineSize: page.contentWidth,
    rowGap: space[40],
  },
  link: {
    color: { default: "inherit", ":hover": color.bandHover },
    outlineColor: { default: null, ":focus-visible": color.bandHover },
    outlineOffset: { default: null, ":focus-visible": border.focusOffset },
    outlineStyle: { default: null, ":focus-visible": "solid" },
    outlineWidth: { default: null, ":focus-visible": border.focusRing },
    textDecorationLine: { default: "none", ":hover": "underline" },
    textDecorationThickness: decoration.hover,
    textUnderlineOffset: decoration.offset,
  },
  brand: { gap: space[16], alignContent: "start", display: "grid", justifyItems: "start", },
  lockup: { gap: space[10], alignItems: "center", display: "inline-flex", textDecorationLine: "none", },
  credit: { margin: 0, color: color.bandMuted, fontSize: fontSize[14], },
  social: { margin: 0, padding: 0, gap: space[8], listStyle: "none", display: "flex", },
  socialLink: {
    borderColor: { default: color.bandRule, ":hover": color.bandAccent },
    borderStyle: "solid",
    borderWidth: border.static,
    placeItems: "center",
    backgroundColor: { default: null, ":hover": color.bandHoverWash },
    blockSize: size[40],
    boxSizing: "border-box",
    color: { default: color.bandText, ":hover": color.bandHover },
    display: "grid",
    inlineSize: size[40],
  },
  socialIcon: { blockSize: size[20], inlineSize: size[20] },
  /** Equal columns under small uppercase labels, so the links lead rather than the headings. */
  groups: {
    columnGap: { default: space[40], [breakpoints.narrow]: space[24] },
    display: "grid",
    gridTemplateColumns: {
      default: `repeat(3, minmax(0, ${size[160]}))`,
      [breakpoints.narrow]: `repeat(auto-fill, minmax(${size[144]}, 1fr))`,
    },
    rowGap: { default: space[32], [breakpoints.narrow]: space[28] },
  },
  groupTitle: {
    marginInline: 0,
    color: color.bandLabel,
    fontFamily: font.body,
    fontSize: fontSize[12],
    fontWeight: fontWeight.semibold,
    letterSpacing: tracking.widest,
    lineHeight: lineHeight.compact,
    marginBlockEnd: space[14],
    marginBlockStart: 0,
    textTransform: "uppercase",
  },
  groupLinks: {
    margin: 0,
    padding: 0,
    gap: space[8],
    listStyle: "none",
    color: color.bandLink,
    display: "grid",
    fontSize: fontSize[14],
    lineHeight: lineHeight.normal,
  },
});

/** Styles for links inside `credit`, matching the band. */
export const siteFooterCreditLink = stylex.create({
  link: {
    color: { default: color.bandText, ":hover": color.bandHover },
    outlineColor: { default: null, ":focus-visible": color.bandHover },
    outlineOffset: { default: null, ":focus-visible": border.focusOffset },
    outlineStyle: { default: null, ":focus-visible": "solid" },
    outlineWidth: { default: null, ":focus-visible": border.focusRing },
    textDecorationColor: { default: color.bandAccent, ":hover": "currentColor" },
    textDecorationLine: "underline",
    textDecorationThickness: { default: null, ":hover": decoration.hover },
    textUnderlineOffset: decoration.offset,
  },
}).link;

/** The site footer: the brand, a credit line and community links beside grouped page links. */
export function SiteFooter({
  brand,
  brandHref,
  brandLabel,
  credit,
  social = [],
  socialLabel = "Community",
  groups,
  label = "Footer",
  renderLink,
  xstyle,
}: SiteFooterProps) {
  return (
    <footer {...stylex.props(styles.band, xstyle)}>
      <div {...stylex.props(styles.inner)}>
        <div {...stylex.props(styles.brand)}>
          {renderNavigationLink(renderLink, {
            href: brandHref,
            "aria-label": brandLabel,
            ...stylex.props(styles.link, styles.lockup),
            children: brand,
          })}
          {credit ? <p {...stylex.props(styles.credit)}>{credit}</p> : null}
          {social.length ? (
            <ul {...stylex.props(styles.social)} aria-label={socialLabel}>
              {social.map((item) => (
                <li key={item.href}>
                  <a href={item.href} aria-label={item.label} {...stylex.props(styles.link, styles.socialLink)}>
                    <span aria-hidden="true" {...stylex.props(styles.socialIcon)}>{fillIcon(item.icon)}</span>
                  </a>
                </li>
              ))}
            </ul>
          ) : null}
        </div>
        <nav {...stylex.props(styles.groups)} aria-label={label}>
          {groups.map((group) => (
            <div key={group.title}>
              <h2 {...stylex.props(styles.groupTitle)}>{group.title}</h2>
              <ul {...stylex.props(styles.groupLinks)}>
                {group.links.map((link) => (
                  <li key={link.id}>
                    {renderNavigationLink(renderLink, {
                      href: link.href,
                      ...stylex.props(styles.link),
                      children: link.label,
                    })}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </nav>
      </div>
    </footer>
  );
}
