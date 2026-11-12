import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
} from "react";
import { XMarkIcon, iconClassName } from "../../../icons";
import { buttonClassName } from "../../actions/Button";
import { surface } from "../../../foundations/recipes";
import { cn, tv } from "../../../foundations/variants";

const drawer = tv({
  slots: {
    backdrop: [
      "paper-drawer__backdrop fixed inset-0 z-backdrop bg-scrim opacity-full transition-opacity duration-standard",
      "data-starting-style:opacity-0 data-ending-style:opacity-0 motion-reduce:transition-none",
      "forced-colors:bg-forced-scrim forced-colors:forced-color-adjust-none",
    ],
    viewport: "paper-drawer__viewport pointer-events-none fixed inset-0 z-dialog flex",
    popup: [
      "paper-drawer pointer-events-auto flex h-full w-drawer flex-col overflow-hidden",
      "font-body text-body leading-body text-ink transition-transform duration-standard motion-reduce:transition-none",
      "forced-colors:border-forced-ink",
    ],
    header: [
      "paper-drawer__header flex flex-none items-start justify-between gap-inline border-b border-rule-subtle p-field-x",
    ],
    heading: "paper-drawer__heading min-w-0",
    title: "paper-drawer__title m-0 font-display text-section leading-section font-bold",
    description: "paper-drawer__description mb-0 mt-1.5 text-muted",
    close: "paper-drawer__close w-control flex-none px-0",
    body: "paper-drawer__body min-h-0 flex-auto overflow-y-auto overscroll-contain p-field-x",
    footer: "paper-drawer__footer flex flex-none justify-end gap-inline border-t border-rule-subtle p-field-x",
  },
  variants: {
    placement: {
      start: {
        viewport: "justify-start",
        popup: [
          "border-y-0 border-s-0 data-starting-style:-translate-x-full data-ending-style:-translate-x-full",
          "rtl:data-starting-style:translate-x-full rtl:data-ending-style:translate-x-full",
        ],
      },
      end: {
        viewport: "justify-end",
        popup: [
          "border-y-0 border-e-0 data-starting-style:translate-x-full data-ending-style:translate-x-full",
          "rtl:data-starting-style:-translate-x-full rtl:data-ending-style:-translate-x-full",
        ],
      },
    },
    size: {
      default: {},
      full: {
        popup: "max-tablet:w-full max-tablet:border-0 max-tablet:shadow-none",
      },
    },
  },
});

export const DRAWER_PLACEMENTS = ["start", "end"] as const;

export type DrawerPlacement = (typeof DRAWER_PLACEMENTS)[number];

export const DRAWER_SIZES = ["default", "full"] as const;

export type DrawerSize = (typeof DRAWER_SIZES)[number];

export interface DrawerProps {
  /** The application-owned control that opens the drawer. */
  readonly trigger: ReactElement;
  /** A visible title that supplies the dialog's accessible name. */
  readonly title: ReactNode;
  /** Optional supporting copy associated with the dialog. */
  readonly description?: ReactNode;
  /** Application-owned drawer body. */
  readonly children: ReactNode;
  /** Optional application-owned region below the scrolling body. */
  readonly footer?: ReactNode;
  readonly placement?: DrawerPlacement;
  /** `full` covers the whole viewport below the tablet breakpoint, for tasks with long lists such as filters. */
  readonly size?: DrawerSize;
  readonly closeLabel?: string;
  readonly defaultOpen?: boolean;
  readonly open?: boolean;
  readonly onOpenChange?: (open: boolean) => void;
}

/**
 * A modal side sheet for supporting tasks and compact navigation.
 *
 * Focus containment, background interaction blocking, Escape dismissal, outside
 * press dismissal, and focus restoration are delegated to Base UI's Dialog.
 */
export function Drawer({
  trigger,
  title,
  description,
  children,
  footer,
  placement = "end",
  size = "default",
  closeLabel = "Close",
  defaultOpen,
  open,
  onOpenChange,
}: DrawerProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const slots = drawer({ placement, size });

  return (
    <Dialog.Root
      defaultOpen={defaultOpen}
      modal
      open={open}
      onOpenChange={(nextOpen) => onOpenChange?.(nextOpen)}
    >
      <Dialog.Trigger
        render={trigger}
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
      />
      <Dialog.Portal container={portalContainer}>
        <Dialog.Backdrop className={slots.backdrop()} />
        <Dialog.Viewport
          className={slots.viewport()}
          data-placement={placement}
        >
          <Dialog.Popup
            className={cn(surface({ tone: "overlay", elevation: "showcase" }), slots.popup())}
            data-placement={placement}
            data-size={size}
          >
            <header className={slots.header()}>
              <div className={slots.heading()}>
                <Dialog.Title className={slots.title()}>
                  {title}
                </Dialog.Title>
                {description ? (
                  <Dialog.Description className={slots.description()}>
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                className={buttonClassName({
                  variant: "ghost",
                  className: slots.close(),
                })}
                aria-label={closeLabel}
              >
                <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div className={slots.body()}>{children}</div>
            {footer ? <footer className={slots.footer()}>{footer}</footer> : null}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
