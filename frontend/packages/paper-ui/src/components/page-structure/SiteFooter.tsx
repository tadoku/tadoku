import type { ReactNode } from "react";
import { tv } from "../../foundations/variants";
import { renderNavigationLink, type NavigationLinkRenderer } from "../navigation/navigation";

export interface SiteFooterLink {
  readonly label: string;
  readonly href: string;
  readonly icon: ReactNode;
}

export interface SiteFooterGroup {
  readonly title: string;
  readonly links: readonly { readonly label: string; readonly href: string }[];
}

export interface SiteFooterProps {
  /** The brand lockup's images, linked to `brandHref` and named by `brandLabel`. */
  readonly brand: ReactNode;
  readonly brandHref: string;
  readonly brandLabel: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly credit?: ReactNode;
  /** Icon-only community links; each label names its destination. */
  readonly social?: readonly SiteFooterLink[];
  readonly socialLabel?: string;
  readonly groups: readonly SiteFooterGroup[];
  readonly label?: string;
}

const siteFooter = tv({
  slots: {
    root: [
      "paper-site-footer bg-footer px-gutter pb-14 pt-12 text-footer-ink",
      "[&_a]:text-inherit [&_a]:no-underline [&_a:hover]:text-footer-hover [&_a:hover]:underline",
      "[&_a:hover]:decoration-hover [&_a:hover]:underline-offset-link [&_:focus-visible]:outline-footer-hover",
      "[&_p_a]:text-footer-ink [&_p_a]:underline [&_p_a]:decoration-footer-accent [&_p_a]:underline-offset-link",
      "[&_p_a:hover]:text-footer-ink",
    ],
    inner: [
      "mx-auto grid max-w-content grid-cols-content-end gap-x-16 gap-y-10 max-lg:grid-cols-single",
    ],
    brand: "grid content-start justify-items-start gap-4",
    lockup: "inline-flex items-center gap-2.5 [&:hover]:no-underline",
    credit: "m-0 text-sm text-footer-muted",
    social: "m-0 flex list-none gap-2 p-0",
    socialLink: [
      "grid size-10 place-items-center border border-footer-rule text-footer-ink [&_svg]:size-icon",
      "hover:border-footer-accent hover:bg-footer-tint hover:text-footer-hover",
    ],
    links: [
      "grid grid-cols-footer-links gap-x-10 gap-y-8",
      "max-tablet:grid-cols-footer-links-narrow max-tablet:gap-x-6 max-tablet:gap-y-7",
    ],
    groupTitle: "mb-3.5 mt-0 font-body text-xs leading-140 font-semibold tracking-caps text-footer-label uppercase",
    groupList: "m-0 grid list-none gap-2 p-0 text-sm leading-150 text-footer-link",
  },
});

const slots = siteFooter();

/** The site's closing band: the brand, a credit line and community links beside grouped page links. */
export function SiteFooter({
  brand,
  brandHref,
  brandLabel,
  renderLink,
  credit,
  social = [],
  socialLabel = "Community",
  groups,
  label = "Footer",
}: SiteFooterProps) {
  return (
    <footer className={slots.root()}>
      <div className={slots.inner()}>
        <div className={slots.brand()}>
          {renderNavigationLink(renderLink, {
            href: brandHref,
            className: slots.lockup(),
            "aria-label": brandLabel,
            children: brand,
          })}
          {credit ? <p className={slots.credit()}>{credit}</p> : null}
          {social.length ? (
            <ul className={slots.social()} aria-label={socialLabel}>
              {social.map((link) => (
                <li key={link.href}>
                  <a className={slots.socialLink()} href={link.href} aria-label={link.label}>{link.icon}</a>
                </li>
              ))}
            </ul>
          ) : null}
        </div>
        <nav className={slots.links()} aria-label={label}>
          {groups.map((group) => (
            <div key={group.title}>
              <h2 className={slots.groupTitle()}>{group.title}</h2>
              <ul className={slots.groupList()}>
                {group.links.map((link) => (
                  <li key={link.href}>{renderNavigationLink(renderLink, { href: link.href, children: link.label })}</li>
                ))}
              </ul>
            </div>
          ))}
        </nav>
      </div>
    </footer>
  );
}
