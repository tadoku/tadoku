import type { MouseEvent, ReactNode } from "react";
import { ChevronLeftIcon, ChevronRightIcon, EllipsisHorizontalIcon, iconClassName } from "../../icons";
import { renderNavigationLink, type NavigationLinkRenderer } from "./navigation";
import { pagination, type PaginationSlot } from "./styles";

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
  readonly className?: string;
  /** Classes merged over a part's own, for a screen's layout needs. */
  readonly classNames?: Partial<Record<PaginationSlot, string>>;
}

const slots = pagination();

export function Pagination({
  totalPages,
  currentPage,
  getHref,
  onPageChange,
  siblingCount = 2,
  label = "Pagination",
  renderLink,
  className,
  classNames = {},
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
      return <span className={slots.control()} aria-disabled="true">{content}</span>;
    }
    if (getHref) {
      return renderNavigationLink(renderLink, {
        href: getHref(page),
        className: slots.control(),
        "aria-label": controlLabel,
        onClick: activate(page),
        children: content,
      });
    }
    return <button type="button" className={slots.control()} aria-label={controlLabel} onClick={() => onPageChange?.(page)}>{content}</button>;
  };

  return (
    <nav className={slots.root({ className: [classNames.root, className] })} aria-label={label}>
      <ul className={slots.list({ className: classNames.list })}>
        <li>{control(currentPage - 1, <><ChevronLeftIcon className={iconClassName("compact")} aria-hidden="true" /><span>Previous</span></>, "Previous page", currentPage === 1)}</li>
        {tokens.map((token) =>
          typeof token === "string" ? (
            <li key={token} className={slots.gap()} aria-hidden="true"><EllipsisHorizontalIcon className={iconClassName("default")} /></li>
          ) : (
            <li key={token} className={token === currentPage ? undefined : slots.otherPage()}>
              {token === currentPage ? (
                <span className={slots.page()} aria-current="page" aria-label={`Page ${token} of ${totalPages}`}>
                  {token}<span className={slots.of()} aria-hidden="true">&nbsp;of {totalPages}</span>
                </span>
              ) : getHref ? renderNavigationLink(renderLink, {
                href: getHref(token),
                className: slots.page(),
                "aria-label": `Page ${token}`,
                onClick: activate(token),
                children: token,
              }) : (
                <button type="button" className={slots.page()} aria-label={`Page ${token}`} onClick={() => onPageChange?.(token)}>{token}</button>
              )}
            </li>
          ),
        )}
        <li>{control(currentPage + 1, <><span>Next</span><ChevronRightIcon className={iconClassName("compact")} aria-hidden="true" /></>, "Next page", currentPage === totalPages)}</li>
      </ul>
    </nav>
  );
}
