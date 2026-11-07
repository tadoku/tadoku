import * as stylex from "@stylexjs/stylex";
import { Dialog } from "@base-ui/react/dialog";
import {
  useState,
  type ReactElement,
  type ReactNode,
} from "react";
import { XMarkIcon } from "../../../icons";
import { buttonStyles } from "../../actions/Button";
import { elevation, icon, surface } from "../../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  font,
  fontWeight,
  motion,
  size,
  space,
  type,
  zIndex,
} from "../../../tokens/tokens.stylex";
import { backdrop } from "../overlay-styles";

const styles = stylex.create({
  viewport: { inset: 0, display: "flex", pointerEvents: "none", position: "fixed", zIndex: zIndex.dialog, },
  start: { justifyContent: "flex-start" },
  end: { justifyContent: "flex-end" },
  popup: {
    borderColor: { default: color.ruleStrong, [breakpoints.forcedColors]: "CanvasText" },
    borderStyle: "solid",
    borderWidth: border.static,
    overflow: "hidden",
    backgroundColor: color.surfaceOverlay,
    blockSize: "100%",
    borderBlockWidth: 0,
    boxSizing: "border-box",
    color: color.textInk,
    display: "flex",
    flexDirection: "column",
    fontFamily: font.body,
    fontSize: density.bodySize,
    inlineSize: `min(${size[448]}, calc(100vw - ${density.controlHeight}))`,
    lineHeight: density.bodyLineHeight,
    pointerEvents: "auto",
    transitionDuration: motion.standard,
    transitionProperty: { default: "transform", [breakpoints.reducedMotion]: "none" },
  },
  popupStart: {
    borderInlineStartWidth: 0,
    transform: { "[data-ending-style]": { default: "translateX(-100%)", ":dir(rtl)": "translateX(100%)" }, "[data-starting-style]": { default: "translateX(-100%)", ":dir(rtl)": "translateX(100%)" }, default: "translateX(0)", },
  },
  popupEnd: {
    borderInlineEndWidth: 0,
    transform: { "[data-ending-style]": { default: "translateX(100%)", ":dir(rtl)": "translateX(-100%)" }, "[data-starting-style]": { default: "translateX(100%)", ":dir(rtl)": "translateX(-100%)" }, default: "translateX(0)", },
  },
  header: {
    padding: density.fieldPaddingInline,
    flex: "none",
    gap: density.inlineGap,
    alignItems: "flex-start",
    borderBlockEndColor: color.ruleSubtle,
    borderBlockEndStyle: "solid",
    borderBlockEndWidth: border.static,
    display: "flex",
    justifyContent: "space-between",
  },
  heading: { minInlineSize: 0 },
  title: {
    margin: 0,
    fontFamily: font.display,
    fontSize: type.sectionSize,
    fontWeight: fontWeight.bold,
    lineHeight: type.sectionLineHeight,
  },
  description: { marginInline: 0, color: color.textMuted, marginBlockEnd: 0, marginBlockStart: space[6], },
  close: { flex: "none", paddingInline: 0, inlineSize: density.controlHeight, },
  body: {
    padding: density.fieldPaddingInline,
    overscrollBehavior: "contain",
    flexBasis: 'auto',
    flexGrow: '1',
    flexShrink: '1',
    minBlockSize: 0,
    overflowY: "auto",
  },
  footer: {
    padding: density.fieldPaddingInline,
    flex: "none",
    gap: density.inlineGap,
    borderBlockStartColor: color.ruleSubtle,
    borderBlockStartStyle: "solid",
    borderBlockStartWidth: border.static,
    display: "flex",
    justifyContent: "flex-end",
  },
});

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
        <Dialog.Backdrop {...stylex.props(backdrop.root)} data-backdrop="" />
        <Dialog.Viewport
          {...stylex.props(styles.viewport, styles[placement])}
          data-placement={placement}
        >
          <Dialog.Popup
            {...stylex.props(
              surface.raised,
              elevation.showcase,
              styles.popup,
              placement === "start" ? styles.popupStart : styles.popupEnd,
            )}
            data-placement={placement}
          >
            <header {...stylex.props(styles.header)}>
              <div {...stylex.props(styles.heading)}>
                <Dialog.Title {...stylex.props(styles.title)}>
                  {title}
                </Dialog.Title>
                {description ? (
                  <Dialog.Description {...stylex.props(styles.description)}>
                    {description}
                  </Dialog.Description>
                ) : null}
              </div>
              <Dialog.Close
                {...stylex.props(...buttonStyles({ variant: "ghost" }), styles.close)}
                aria-label={closeLabel}
              >
                <XMarkIcon {...stylex.props(icon.prominent)} aria-hidden="true" />
              </Dialog.Close>
            </header>
            <div {...stylex.props(styles.body)}>{children}</div>
            {footer ? <footer {...stylex.props(styles.footer)}>{footer}</footer> : null}
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
