import type { AnchorHTMLAttributes } from "react";
import { cn } from "../../foundations/variants";
import { buttonClassName } from "./Button";

/** The first focusable element of a page: hidden above the viewport until it receives focus. */
export function SkipLink({ className, ...props }: AnchorHTMLAttributes<HTMLAnchorElement>) {
  return (
    <a
      {...props}
      className={buttonClassName({
        className: cn("paper-skip-link fixed start-3 top-3 z-skip -translate-y-offscreen focus:translate-y-0", className),
      })}
    />
  );
}
