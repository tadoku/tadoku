import * as stylex from "@stylexjs/stylex";
import { useId } from "react";
import { IconSlot } from "../../foundations/IconSlot";
import { a11y } from "../../foundations/recipes";
import { list, sidebar } from "./navigation-styles";
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
    <nav {...stylex.props(sidebar.root)} aria-label={label}>
      {sections.map((section) => (
        <section key={section.id} {...stylex.props(sidebar.section)} aria-labelledby={sidebarSectionHeadingId(sidebarId, section.id)}>
          <h2 id={sidebarSectionHeadingId(sidebarId, section.id)} {...stylex.props(sidebar.title)}>{section.title}</h2>
          <ul {...stylex.props(list.reset, sidebar.section)}>
            {section.links.map((link) => {
              const current = linkIsCurrent(link, currentPath);
              return (
                <li key={link.id}>
                  {renderNavigationLink(renderLink, {
                    href: link.href,
                    ...stylex.props(sidebar.link, a11y.focusRing),
                    "aria-current": current ? "page" : undefined,
                    "aria-disabled": link.disabled || undefined,
                    tabIndex: link.disabled ? -1 : undefined,
                    onClick: link.disabled ? disabledLinkClick : () => link.onSelect?.(),
                    children: (
                      <>
                        {link.icon ? <IconSlot>{link.icon}</IconSlot> : null}
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
