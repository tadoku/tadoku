import { useId, useState } from "react";
import { focusRing } from "../../foundations/recipes";
import { cn } from "../../foundations/variants";

export interface SpoilerProps {
  /** The text to keep blurred, and hidden from assistive technology, until the reader reveals it. */
  readonly children: string;
  /** Names the text in the button's label, such as "NSFW description" for "Reveal NSFW description". */
  readonly label: string;
  /** A short word shown over the blurred text, such as "NSFW". */
  readonly badge: string;
  readonly className?: string;
}

/** Text the reader reveals with a click, or previews while a mouse rests on it. Before hydration it stays hidden. */
export function Spoiler({ children, label, badge, className }: SpoilerProps) {
  const id = useId();
  const [revealed, setRevealed] = useState(false);
  const [hovered, setHovered] = useState(false);
  const visible = revealed || hovered;

  return (
    <button
      type="button"
      className={cn(
        "paper-spoiler relative m-0 cursor-pointer border-0 bg-transparent p-0 text-start text-inherit",
        focusRing(),
        className,
      )}
      aria-label={`${revealed ? "Hide" : "Reveal"} ${label}`}
      aria-expanded={visible}
      aria-describedby={visible ? id : undefined}
      onPointerEnter={(event) => {
        if (event.pointerType === "mouse") setHovered(true);
      }}
      onPointerLeave={() => setHovered(false)}
      onClick={() => setRevealed(!revealed)}
    >
      <span id={id} aria-hidden={!visible} className={cn("pointer-events-none", !visible && "select-none blur-spoiler")}>
        {children}
      </span>
      {visible ? null : (
        <span
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 flex items-center justify-center text-fine font-bold"
        >
          {badge}
        </span>
      )}
    </button>
  );
}
