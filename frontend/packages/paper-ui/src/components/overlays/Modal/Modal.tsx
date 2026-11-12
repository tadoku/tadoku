import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
  type RefObject,
} from "react";
import { XMarkIcon, iconClassName } from "../../../icons";
import { buttonClassName, type ButtonVariant } from "../../actions/Button";
import { surface } from "../../../foundations/recipes";
import { cn, tv } from "../../../foundations/variants";

const modal = tv({
  slots: {
    backdrop: [
      "paper-modal__backdrop fixed inset-0 z-backdrop bg-scrim opacity-full backdrop-blur-scrim",
      "transition-opacity duration-standard data-starting-style:opacity-0 data-ending-style:opacity-0",
      "motion-reduce:transition-none forced-colors:bg-forced-scrim forced-colors:backdrop-blur-none",
      "forced-colors:forced-color-adjust-none",
    ],
    viewport: "paper-modal__viewport fixed inset-0 z-dialog grid place-items-center overflow-y-auto p-4",
    popup: [
      "paper-modal max-h-modal w-modal overflow-y-auto transition-enter duration-standard motion-reduce:transition-none",
      "data-starting-style:translate-y-2 data-starting-style:opacity-0",
      "data-ending-style:translate-y-2 data-ending-style:opacity-0",
    ],
    header: "paper-modal__header flex items-start justify-between gap-4 border-b border-rule-subtle p-5",
    title: "paper-modal__title m-0 font-display text-section leading-section",
    description: "paper-modal__description mb-0 mt-1.6 text-muted",
    close: [
      "paper-modal__icon-close grid size-control flex-none cursor-pointer place-items-center",
      "border border-transparent bg-transparent p-0 text-2xl text-muted",
      "hover:border-rule hover:bg-raised hover:text-ink",
    ],
    content: "paper-modal__content p-5",
    footer: "paper-modal__footer flex justify-end gap-inline border-t border-rule-subtle px-5 py-4",
  },
});

const slots = modal();
const popupClassName = cn(surface({ tone: "overlay", elevation: "showcase" }), slots.popup());

type ModalInteractionType = "mouse" | "touch" | "pen" | "keyboard" | "";
type ModalInitialFocus =
  | boolean
  | RefObject<HTMLElement | null>
  | ((openType: ModalInteractionType) => boolean | HTMLElement | null | void);

interface ModalContentProps {
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

type ModalTrigger =
  | {
      /** Label for the standard Paper button trigger. */
      readonly triggerLabel: string;
      readonly triggerVariant?: ButtonVariant;
      readonly trigger?: never;
    }
  | {
      /**
       * Application-owned trigger element. Paper adds the dialog behavior and ARIA
       * state without replacing the element's appearance or ref.
       */
      readonly trigger: ReactElement;
      readonly triggerLabel?: never;
      readonly triggerVariant?: never;
    }
  | {
      /**
       * No trigger: another control, such as an `ActionMenu` item, opens the
       * dialog through the controlled `open` state.
       */
      readonly trigger?: never;
      readonly triggerLabel?: never;
      readonly triggerVariant?: never;
      readonly open: boolean;
      readonly onOpenChange: (open: boolean) => void;
    };

export type ModalProps = ModalContentProps & ModalTrigger;

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
      ) : triggerLabel ? (
        <Dialog.Trigger
          ref={captureTrigger}
          className={buttonClassName({ variant: triggerVariant })}
        >
          {triggerLabel}
        </Dialog.Trigger>
      ) : null}
      {/* A trigger names the document to portal into, such as a catalogue iframe; without one, the current document. */}
      <Dialog.Portal container={trigger || triggerLabel ? portalContainer : undefined}>
        <Dialog.Backdrop className={slots.backdrop()} />
        <Dialog.Viewport className={slots.viewport()}>
          <Dialog.Popup
            className={popupClassName}
            initialFocus={initialFocus}
          >
            <header className={slots.header()}>
              <div>
                <Dialog.Title className={slots.title()}>{title}</Dialog.Title>
                {description ? (
                  <Dialog.Description className={slots.description()}>
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                className={slots.close()}
                aria-label={closeLabel}
              >
                <XMarkIcon className={iconClassName("prominent")} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div className={slots.content()}>{children}</div>
            {footer === undefined ? (
              <footer className={slots.footer()}>
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
              <footer className={slots.footer()}>{footer}</footer>
            )}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
