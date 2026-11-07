import type { HTMLAttributes } from "react";
import { cn } from "../../foundations/variants";

/** A decorative icon on a soft accent tile, leading a step, feature or empty list. */
export function IconTile({ className, ...props }: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      aria-hidden="true"
      {...props}
      className={cn("paper-icon-tile grid size-target flex-none place-items-center bg-action-soft text-link", className)}
    />
  );
}
