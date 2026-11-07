import * as stylex from "@stylexjs/stylex";
import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
  type RefObject,
} from "react";
import { XMarkIcon } from "../../../icons";
import { buttonStyles, type ButtonVariant } from "../../actions/Button";
import { elevation, icon, surface } from "../../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  font,
  fontSize,
  motion,
  size,
  space,
  type,
  zIndex,
} from "../../../tokens/tokens.stylex";
import { backdrop } from "../overlay-styles";

const styles = stylex.create({
  viewport: {
    inset: 0,
    padding: space[16],
    placeItems: "center",
    display: "grid",
    position: "fixed",
    zIndex: zIndex.dialog,
    overflowY: "auto",
  },
  popup: {
    borderColor: color.ruleStrong,
    borderStyle: "solid",
    borderWidth: border.static,
    backgroundColor: color.surfaceOverlay,
    boxSizing: "border-box",
    inlineSize: `min(${size[576]}, 100%)`,
    maxBlockSize: `min(${size[672]}, calc(100vh - ${space[32]}))`,
    opacity: { "[data-ending-style]": 0, "[data-starting-style]": 0, default: null, },
    transform: {
      "[data-ending-style]": `translateY(${space[8]})`,
      "[data-starting-style]": `translateY(${space[8]})`,
      default: "translateY(0)",
    },
    transitionDuration: motion.standard,
    transitionProperty: { default: "opacity, transform", [breakpoints.reducedMotion]: "none" },
    overflowY: "auto",
  },
  header: {
    padding: space[20],
    gap: space[16],
    alignItems: "start",
    borderBlockEndColor: color.ruleSubtle,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    display: "flex",
    justifyContent: "space-between",
  },
  title: {
    margin: 0,
    fontFamily: font.display,
    fontSize: type.sectionSize,
    lineHeight: type.sectionLineHeight,
  },
  description: { marginInline: 0, color: color.textMuted, marginBlockEnd: 0, marginBlockStart: space[6], },
  close: {
    padding: 0,
    borderColor: { default: "transparent", ":hover": color.ruleDefault },
    borderStyle: "solid",
    borderWidth: border.static,
    flex: "none",
    placeItems: "center",
    backgroundColor: { default: "transparent", ":hover": color.surfaceRaised },
    blockSize: density.controlHeight,
    boxSizing: "border-box",
    color: { default: color.textMuted, ":hover": color.textInk },
    cursor: "pointer",
    display: "grid",
    fontSize: fontSize[24],
    inlineSize: density.controlHeight,
    outlineColor: {
      default: null,
      ":focus-visible": color.focusRing,
    },
    outlineOffset: { default: null, ":focus-visible": border.focusOffset },
    outlineStyle: {
      default: null,
      ":focus-visible": "solid",
    },
    outlineWidth: {
      default: null,
      ":focus-visible": border.focusRing,
    },
  },
  content: { padding: space[20] },
  footer: {
    gap: density.inlineGap,
    paddingBlock: space[16],
    paddingInline: space[20],
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "flex",
    justifyContent: "end",
  },
});

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
          {...stylex.props(...buttonStyles({ variant: triggerVariant }))}
        >
          {triggerLabel}
        </Dialog.Trigger>
      )}
      <Dialog.Portal container={portalContainer}>
        <Dialog.Backdrop {...stylex.props(backdrop.root, backdrop.blurred)} data-backdrop="" />
        <Dialog.Viewport {...stylex.props(styles.viewport)}>
          <Dialog.Popup
            {...stylex.props(surface.raised, elevation.showcase, styles.popup)}
            initialFocus={initialFocus}
          >
            <header {...stylex.props(styles.header)}>
              <div>
                <Dialog.Title {...stylex.props(styles.title)}>{title}</Dialog.Title>
                {description ? (
                  <Dialog.Description {...stylex.props(styles.description)}>
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                {...stylex.props(styles.close)}
                aria-label={closeLabel}
              >
                <XMarkIcon {...stylex.props(icon.prominent)} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div {...stylex.props(styles.content)}>{children}</div>
            {footer === undefined ? (
              <footer {...stylex.props(styles.footer)}>
                {action ? (
                  <Dialog.Close
                    {...stylex.props(...buttonStyles({ variant: action.variant }))}
                    disabled={action.disabled}
                    onClick={action.onAction}
                  >
                    {action.label}
                  </Dialog.Close>
                ) : null}
                <Dialog.Close {...stylex.props(...buttonStyles({ variant: "outline" }))}>
                  {closeLabel}
                </Dialog.Close>
              </footer>
            ) : footer === null ? null : (
              <footer {...stylex.props(styles.footer)}>{footer}</footer>
            )}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
