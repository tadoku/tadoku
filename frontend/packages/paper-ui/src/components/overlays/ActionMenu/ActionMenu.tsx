import { Menu } from "@base-ui/react/menu";
import { useState, type ReactNode } from "react";
import { buttonClassName, type ButtonVariant } from "../../actions/Button";
import { ChevronDownIcon, EllipsisHorizontalIcon, iconClassName } from "../../../icons";
import { surface } from "../../../foundations/recipes";
import { cn, tv } from "../../../foundations/variants";

const actionMenu = tv({
  slots: {
    iconOnly: "paper-action-menu__trigger--icon-only w-target min-h-target min-w-target shrink-0 p-0",
    positioner: "paper-action-menu__positioner z-menu outline-none",
    popup: [
      "paper-action-menu min-w-menu origin-(--transform-origin) py-1.4",
      "transition-enter duration-quick motion-reduce:transition-none",
      "data-starting-style:scale-enter data-starting-style:opacity-0",
      "data-ending-style:scale-enter data-ending-style:opacity-0",
    ],
    item: [
      "paper-action-menu__item flex min-h-control cursor-default select-none items-center gap-inline",
      "px-field-x py-field-y outline-none data-highlighted:bg-action data-highlighted:text-on-action",
      "data-disabled:opacity-faded",
    ],
    destructive: [
      "paper-action-menu__item--destructive text-danger data-highlighted:bg-destructive",
      "data-highlighted:text-inverse",
    ],
    icon: "paper-action-menu__icon grid size-icon place-items-center *:size-full",
    heading: [
      "paper-action-menu__heading mb-1 border-b border-rule-subtle px-field-x pb-2 pt-1",
      "text-sm font-semibold text-muted wrap-anywhere",
    ],
  },
});

const slots = actionMenu();
const popupClassName = cn(surface({ tone: "overlay", elevation: "floating" }), slots.popup());

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
  /** A label above the items that names what they act on, such as the signed-in account. */
  readonly heading?: string;
}

export function ActionMenu({
  label,
  items,
  triggerVariant = "outline",
  defaultOpen,
  iconOnly = false,
  triggerIcon,
  heading,
}: ActionMenuProps) {
  const [portalContainer, setPortalContainer] = useState<HTMLElement | null>(null);
  const menuItems = items.map((item) => (
    <Menu.Item
      key={item.id}
      className={slots.item({ className: item.destructive ? slots.destructive() : undefined })}
      disabled={item.disabled}
      label={item.label}
      onClick={item.onSelect}
    >
      {item.icon ? (
        <span className={slots.icon()} aria-hidden="true">
          {item.icon}
        </span>
      ) : null}
      <span>{item.label}</span>
    </Menu.Item>
  ));

  return (
    <Menu.Root defaultOpen={defaultOpen}>
      <Menu.Trigger
        ref={(node: HTMLElement | null) => {
          if (node && node.ownerDocument.body !== portalContainer) {
            setPortalContainer(node.ownerDocument.body);
          }
        }}
        className={buttonClassName({ variant: triggerVariant, className: iconOnly ? slots.iconOnly() : undefined })}
        aria-label={iconOnly ? label : undefined}
        title={iconOnly ? label : undefined}
      >
        {!iconOnly && <span>{label}</span>}
        {triggerIcon ? <span className={slots.icon()} aria-hidden="true">{triggerIcon}</span> : iconOnly
          ? <EllipsisHorizontalIcon className={iconClassName()} aria-hidden="true" />
          : <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />}
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner
          className={slots.positioner()}
          sideOffset={6}
          align="end"
        >
          <Menu.Popup className={popupClassName}>
            {heading ? (
              <Menu.Group>
                <Menu.GroupLabel className={slots.heading()}>{heading}</Menu.GroupLabel>
                {menuItems}
              </Menu.Group>
            ) : menuItems}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}
