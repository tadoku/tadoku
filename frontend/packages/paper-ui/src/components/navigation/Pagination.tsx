import * as stylex from "@stylexjs/stylex";
import type { MouseEvent, ReactNode } from "react";
import { ChevronLeftIcon, ChevronRightIcon, EllipsisHorizontalIcon } from "../../icons";
import type { StyleXProp } from "../../foundations/props";
import { a11y, control as controlStyles, icon } from "../../foundations/recipes";
import { list, pagination } from "./navigation-styles";

const item = [controlStyles.inherit, pagination.control, a11y.focusRing] as const;
import { renderNavigationLink, type NavigationLinkRenderer } from "./navigation";

type PaginationToken = number | "start-gap" | "end-gap";

function paginationTokens(totalPages: number, currentPage: number, siblingCount: number): PaginationToken[] {
  const visible = new Set([1, totalPages]);
  for (let page = currentPage - siblingCount; page <= currentPage + siblingCount; page += 1) {
    if (page >= 1 && page <= totalPages) visible.add(page);
  }
  const pages = [...visible].sort((left, right) => left - right);
  const tokens: PaginationToken[] = [];
  pages.forEach((page, index) => {
    const previous = pages[index - 1];
    if (previous !== undefined && page - previous > 1) {
      tokens.push(previous === 1 ? "start-gap" : "end-gap");
    }
    tokens.push(page);
  });
  return tokens;
}

export interface PaginationProps {
  readonly totalPages: number;
  readonly currentPage: number;
  readonly getHref?: (page: number) => string;
  readonly onPageChange?: (page: number) => void;
  readonly siblingCount?: number;
  readonly label?: string;
  readonly renderLink?: NavigationLinkRenderer;
  readonly xstyle?: StyleXProp;
  readonly listXstyle?: StyleXProp;
}

export function Pagination({
  totalPages,
  currentPage,
  getHref,
  onPageChange,
  siblingCount = 2,
  label = "Pagination",
  renderLink,
  xstyle,
  listXstyle,
}: PaginationProps) {
  if (!Number.isInteger(totalPages) || totalPages < 1) {
    throw new RangeError("Pagination totalPages must be a positive integer.");
  }
  if (!Number.isInteger(currentPage) || currentPage < 1 || currentPage > totalPages) {
    throw new RangeError("Pagination currentPage must be within totalPages.");
  }
  if (!Number.isInteger(siblingCount) || siblingCount < 0) {
    throw new RangeError("Pagination siblingCount must be a non-negative integer.");
  }
  if (!getHref && !onPageChange) {
    throw new Error("Pagination requires getHref or onPageChange.");
  }

  const activate = (page: number) => (event: MouseEvent<HTMLAnchorElement>) => {
    if (!onPageChange || event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    onPageChange(page);
  };
  const tokens = paginationTokens(totalPages, currentPage, siblingCount);

  const control = (page: number, content: ReactNode, controlLabel: string, disabled: boolean) => {
    if (disabled) {
      return <span {...stylex.props(...item)} aria-disabled="true">{content}</span>;
    }
    if (getHref) {
      return renderNavigationLink(renderLink, {
        href: getHref(page),
        ...stylex.props(...item),
        "aria-label": controlLabel,
        onClick: activate(page),
        children: content,
      });
    }
    return <button type="button" {...stylex.props(...item)} aria-label={controlLabel} onClick={() => onPageChange?.(page)}>{content}</button>;
  };

  return (
    <nav {...stylex.props(xstyle)} aria-label={label}>
      <ul {...stylex.props(list.reset, pagination.list, listXstyle)}>
        <li>{control(currentPage - 1, <><ChevronLeftIcon {...stylex.props(icon.compact)} aria-hidden="true" /><span>Previous</span></>, "Previous page", currentPage === 1)}</li>
        {tokens.map((token) =>
          typeof token === "string" ? (
            <li key={token} {...stylex.props(pagination.gap)} aria-hidden="true"><EllipsisHorizontalIcon {...stylex.props(icon.default)} /></li>
          ) : (
            <li key={token} {...stylex.props(token !== currentPage && pagination.phoneHidden)}>
              {token === currentPage ? (
                <span {...stylex.props(...item, pagination.current)} aria-current="page" aria-label={`Page ${token} of ${totalPages}`}>
                  {token}<span {...stylex.props(pagination.of)} aria-hidden="true">&nbsp;of {totalPages}</span>
                </span>
              ) : getHref ? renderNavigationLink(renderLink, {
                href: getHref(token),
                ...stylex.props(...item),
                "aria-label": `Page ${token}`,
                onClick: activate(token),
                children: token,
              }) : (
                <button type="button" {...stylex.props(...item)} aria-label={`Page ${token}`} onClick={() => onPageChange?.(token)}>{token}</button>
              )}
            </li>
          ),
        )}
        <li>{control(currentPage + 1, <><span>Next</span><ChevronRightIcon {...stylex.props(icon.compact)} aria-hidden="true" /></>, "Next page", currentPage === totalPages)}</li>
      </ul>
    </nav>
  );
}
