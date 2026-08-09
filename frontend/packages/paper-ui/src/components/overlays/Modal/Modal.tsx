import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
  type RefObject,
} from "react";
import { XMarkIcon, iconClassName } from "../../../icons";
import { buttonClassName, type ButtonVariant } from "../../actions/Button";

type ModalInteractionType = "mouse" | "touch" | "pen" | "keyboard" | "";
type ModalInitialFocus =
  | boolean
  | RefObject<HTMLElement | null>
  | ((openType: ModalInteractionType) => boolean | HTMLElement | null | void);

export interface ModalProps {
  /**
   * Label for the standard Paper button trigger. Required when `trigger` is not
   * supplied.
   */
  readonly triggerLabel?: string;
  /**
   * Application-owned trigger element. Paper adds the dialog behavior and ARIA
   * state without replacing the element's appearance or ref.
   */
  readonly trigger?: ReactElement;
  readonly triggerVariant?: ButtonVariant;
  readonly title: string;
  readonly description?: string;
  readonly children: ReactNode;
  readonly closeLabel?: string;
  readonly action?: {
    readonly label: string;
    readonly variant?: ButtonVariant;
    readonly disabled?: boolean;
    readonly onAction: () => void;
  };
  /**
   * Replaces the standard action/close footer. Pass `null` to omit the footer;
   * the header close button and Escape key remain available dismissal paths.
   */
  readonly footer?: ReactNode;
  readonly defaultOpen?: boolean;
  readonly open?: boolean;
  readonly onOpenChange?: (open: boolean) => void;
  /** Determines which application-owned element receives focus when opened. */
  readonly initialFocus?: ModalInitialFocus;
}

export function Modal({
  triggerLabel,
  trigger,
  triggerVariant = "default",
  title,
  description,
  children,
  closeLabel = "Close",
  action,
  footer,
  defaultOpen,
  open,
  onOpenChange,
  initialFocus,
}: ModalProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);

  const captureTrigger = (node: HTMLElement | null) => {
    if (node && node.ownerDocument.body !== portalContainer) {
      setPortalContainer(node.ownerDocument.body);
    }
  };

  return (
    <Dialog.Root
      defaultOpen={defaultOpen}
      open={open}
      onOpenChange={(nextOpen) => onOpenChange?.(nextOpen)}
    >
      {trigger ? (
        <Dialog.Trigger ref={captureTrigger} render={trigger} />
      ) : (
        <Dialog.Trigger
          ref={captureTrigger}
          className={buttonClassName({ variant: triggerVariant })}
        >
          {triggerLabel}
        </Dialog.Trigger>
      )}
      <Dialog.Portal container={portalContainer}>
        <Dialog.Backdrop className="paper-modal__backdrop" />
        <Dialog.Viewport className="paper-modal__viewport">
          <Dialog.Popup
            className="paper-modal paper-surface-raised paper-elevation-showcase"
            initialFocus={initialFocus}
          >
            <header className="paper-modal__header">
              <div>
                <Dialog.Title className="paper-modal__title">{title}</Dialog.Title>
                {description ? (
                  <Dialog.Description className="paper-modal__description">
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                className="paper-modal__icon-close"
                aria-label={closeLabel}
              >
                <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div className="paper-modal__content">{children}</div>
            {footer === undefined ? (
              <footer className="paper-modal__footer">
                {action ? (
                  <Dialog.Close
                    className={buttonClassName({ variant: action.variant })}
                    disabled={action.disabled}
                    onClick={action.onAction}
                  >
                    {action.label}
                  </Dialog.Close>
                ) : null}
                <Dialog.Close className={buttonClassName({ variant: "outline" })}>
                  {closeLabel}
                </Dialog.Close>
              </footer>
            ) : footer === null ? null : (
              <footer className="paper-modal__footer">{footer}</footer>
            )}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
