import { useId } from "react";
import {
  disabledLinkClick,
  linkIsCurrent,
  renderNavigationLink,
  type NavigationItem,
  type NavigationLinkRenderer,
} from "./navigation";

export interface SidebarSection {
  readonly id: string;
  readonly title: string;
  readonly links: readonly NavigationItem[];
}

export interface SidebarProps {
  readonly sections: readonly SidebarSection[];
  readonly currentPath?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly label?: string;
}

function sidebarSectionHeadingId(sidebarId: string, sectionId: string) {
  const suffix = sectionId.replace(/[^a-zA-Z0-9_-]+/gu, "-").replace(/^-+|-+$/gu, "") || "section";
  return `${sidebarId}-${suffix}`;
}

export function Sidebar({
  sections,
  currentPath,
  renderLink,
  label = "Section navigation",
}: SidebarProps) {
  const sidebarId = `paper-sidebar-${useId().replace(/:/gu, "")}`;

  return (
    <nav className="paper-sidebar" aria-label={label}>
      {sections.map((section) => (
        <section key={section.id} className="paper-sidebar__section" aria-labelledby={sidebarSectionHeadingId(sidebarId, section.id)}>
          <h2 id={sidebarSectionHeadingId(sidebarId, section.id)} className="paper-sidebar__title">{section.title}</h2>
          <ul className="paper-sidebar__list">
            {section.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              return (
                <li key={link.id}>
                  {renderNavigationLink(renderLink, {
                    href: link.href,
                    className: "paper-sidebar__link",
                    "aria-current": current ? "page" : undefined,
                    "aria-disabled": link.disabled || undefined,
                    tabIndex: link.disabled ? -1 : undefined,
                    onClick: link.disabled ? disabledLinkClick : () => link.onSelect?.(),
                    children: (
                      <>
                        {link.icon ? <span className="paper-navigation__icon" aria-hidden="true">{link.icon}</span> : null}
                        <span>{link.label}</span>
                      </>
                    ),
                  })}
                </li>
              );
            })}
          </ul>
        </section>
      ))}
    </nav>
  );
}
