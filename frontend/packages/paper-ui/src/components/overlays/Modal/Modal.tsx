import { Dialog } from "@base-ui/react/dialog";
import { useState, type ReactNode } from "react";
import { buttonClassName, type ButtonVariant } from "../../actions/Button";

export interface ModalProps {
  readonly triggerLabel: string;
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
  readonly defaultOpen?: boolean;
  readonly open?: boolean;
  readonly onOpenChange?: (open: boolean) => void;
}

export function Modal({
  triggerLabel,
  triggerVariant = "default",
  title,
  description,
  children,
  closeLabel = "Close",
  action,
  defaultOpen,
  open,
  onOpenChange,
}: ModalProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);

  return (
    <Dialog.Root
      defaultOpen={defaultOpen}
      open={open}
      onOpenChange={(nextOpen) => onOpenChange?.(nextOpen)}
    >
      <Dialog.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
        className={buttonClassName({ variant: triggerVariant })}
      >
        {triggerLabel}
      </Dialog.Trigger>
      <Dialog.Portal container={portalContainer}>
        <Dialog.Backdrop className="paper-modal__backdrop" />
        <Dialog.Viewport className="paper-modal__viewport">
          <Dialog.Popup className="paper-modal paper-surface-raised paper-elevation-showcase">
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
                <span aria-hidden="true">×</span>
              </Dialog.Close>
            </header>
            <div className="paper-modal__content">{children}</div>
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
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
