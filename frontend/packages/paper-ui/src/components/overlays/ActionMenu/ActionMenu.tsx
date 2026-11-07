import * as stylex from "@stylexjs/stylex";
import { Menu } from "@base-ui/react/menu";
import { useState, type ReactNode } from "react";
import { buttonStyles, type ButtonVariant } from "../../actions/Button";
import { ChevronDownIcon, EllipsisHorizontalIcon } from "../../../icons";
import { IconSlot } from "../../../foundations/IconSlot";
import type { StyleXProp } from "../../../foundations/props";
import { elevation, icon } from "../../../foundations/recipes";
import {
  border,
  breakpoints,
  color,
  density,
  motion,
  opacity,
  size,
  space,
  zIndex,
} from "../../../tokens/tokens.stylex";

const styles = stylex.create({
  iconOnly: {
    paddingInline: 0,
    flexShrink: 0,
    inlineSize: size[44],
    minBlockSize: size[44],
    minInlineSize: size[44],
    paddingBlockEnd: 0,
    paddingBlockStart: 0,
  },
  positioner: { outlineStyle: "none", zIndex: zIndex.menu },
  popup: {
    borderColor: color.ruleStrong,
    borderStyle: "solid",
    borderWidth: border.static,
    paddingBlock: space[6],
    backgroundColor: color.surfaceOverlay,
    color: color.textInk,
    minInlineSize: size[208],
    opacity: { "[data-ending-style]": 0, "[data-starting-style]": 0, default: null, },
    transform: { "[data-ending-style]": "scale(0.98)", "[data-starting-style]": "scale(0.98)", default: null, },
    transformOrigin: "var(--transform-origin)",
    transitionDuration: motion.quick,
    transitionProperty: { default: "opacity, transform", [breakpoints.reducedMotion]: "none" },
  },
  item: {
    gap: density.inlineGap,
    paddingBlock: density.fieldPaddingBlock,
    paddingInline: density.fieldPaddingInline,
    alignItems: "center",
    backgroundColor: { "[data-highlighted]": color.actionDefault, default: null, },
    boxSizing: "border-box",
    color: { "[data-highlighted]": color.actionText, default: null, },
    cursor: "default",
    display: "flex",
    minBlockSize: density.controlHeight,
    opacity: { "[data-disabled]": opacity.inactive, default: null, },
    outlineStyle: "none",
    userSelect: "none",
  },
  destructive: {
    backgroundColor: { "[data-highlighted]": color.actionDestructive, default: null, },
    color: { "[data-highlighted]": color.textInverse, default: color.statusDanger, },
  },
});

export interface ActionMenuItem {
  readonly id: string;
  readonly label: string;
  readonly icon?: ReactNode;
  readonly disabled?: boolean;
  readonly destructive?: boolean;
  readonly onSelect: () => void;
}

export interface ActionMenuProps {
  readonly label: string;
  readonly items: readonly ActionMenuItem[];
  readonly triggerVariant?: ButtonVariant;
  readonly defaultOpen?: boolean;
  /** Use the labelled ellipsis trigger when the surrounding item provides context. */
  readonly iconOnly?: boolean;
  /** Replace the default trigger icon; the menu label remains its accessible name. */
  readonly triggerIcon?: ReactNode;
  readonly triggerXstyle?: StyleXProp;
  /** Styles for the trigger's visible label, such as truncating a long account name. */
  readonly triggerLabelXstyle?: StyleXProp;
}

export function ActionMenu({
  label,
  items,
  triggerVariant = "outline",
  defaultOpen,
  iconOnly = false,
  triggerIcon,
  triggerXstyle,
  triggerLabelXstyle,
}: ActionMenuProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);

  return (
    <Menu.Root defaultOpen={defaultOpen}>
      <Menu.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
        {...stylex.props(...buttonStyles({ variant: triggerVariant }), iconOnly && styles.iconOnly, triggerXstyle)}
        aria-label={iconOnly ? label : undefined}
        title={iconOnly ? label : undefined}
      >
        {!iconOnly && <span {...stylex.props(triggerLabelXstyle)}>{label}</span>}
        {triggerIcon ? <IconSlot>{triggerIcon}</IconSlot> : iconOnly
          ? <EllipsisHorizontalIcon {...stylex.props(icon.default)} aria-hidden="true" />
          : <ChevronDownIcon {...stylex.props(icon.compact)} aria-hidden="true" />}
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner
          {...stylex.props(styles.positioner)}
          sideOffset={6}
          align="end"
        >
          <Menu.Popup {...stylex.props(styles.popup, elevation.floating)}>
            {items.map((item) => (
              <Menu.Item
                key={item.id}
                {...stylex.props(styles.item, item.destructive && styles.destructive)}
                disabled={item.disabled}
                label={item.label}
                onClick={item.onSelect}
              >
                {item.icon ? <IconSlot>{item.icon}</IconSlot> : null}
                <span>{item.label}</span>
              </Menu.Item>
            ))}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}
