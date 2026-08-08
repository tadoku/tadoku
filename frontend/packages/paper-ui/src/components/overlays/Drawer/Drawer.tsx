import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
} from "react";
import { XMarkIcon, iconClassName } from "../../../icons";
import { buttonClassName } from "../../actions/Button";

export const DRAWER_PLACEMENTS = ["start", "end"] as const;

export type DrawerPlacement = (typeof DRAWER_PLACEMENTS)[number];

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
  closeLabel = "Close",
  defaultOpen,
  open,
  onOpenChange,
}: DrawerProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);

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
        <Dialog.Backdrop className="paper-drawer__backdrop" />
        <Dialog.Viewport
          className="paper-drawer__viewport"
          data-placement={placement}
        >
          <Dialog.Popup
            className="paper-drawer paper-surface-raised paper-elevation-showcase"
            data-placement={placement}
          >
            <header className="paper-drawer__header">
              <div className="paper-drawer__heading">
                <Dialog.Title className="paper-drawer__title">
                  {title}
                </Dialog.Title>
                {description ? (
                  <Dialog.Description className="paper-drawer__description">
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                className={buttonClassName({
                  variant: "ghost",
                  className: "paper-drawer__close",
                })}
                aria-label={closeLabel}
              >
                <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div className="paper-drawer__body">{children}</div>
            {footer ? <footer className="paper-drawer__footer">{footer}</footer> : null}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
