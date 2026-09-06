import { Menu } from "@base-ui/react/menu";
import { useState, type ReactNode } from "react";
import { buttonClassName, type ButtonVariant } from "../../actions/Button";
import { ChevronDownIcon, EllipsisHorizontalIcon, iconClassName } from "../../../icons";

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
}

export function ActionMenu({
  label,
  items,
  triggerVariant = "outline",
  defaultOpen,
  iconOnly = false,
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
        className={buttonClassName({ variant: triggerVariant })}
        aria-label={iconOnly ? label : undefined}
        title={iconOnly ? label : undefined}
      >
        {iconOnly ? <EllipsisHorizontalIcon className={iconClassName()} aria-hidden="true" /> : <>
          <span>{label}</span>
          <ChevronDownIcon className={iconClassName("compact")} aria-hidden="true" />
        </>}
      </Menu.Trigger>
      <Menu.Portal container={portalContainer}>
        <Menu.Positioner
          className="paper-action-menu__positioner"
          sideOffset={6}
          align="start"
        >
          <Menu.Popup className="paper-action-menu paper-elevation-floating">
            {items.map((item) => (
              <Menu.Item
                key={item.id}
                className={`paper-action-menu__item${
                  item.destructive ? " paper-action-menu__item--destructive" : ""
                }`}
                disabled={item.disabled}
                label={item.label}
                onClick={item.onSelect}
              >
                {item.icon ? (
                  <span className="paper-action-menu__icon" aria-hidden="true">
                    {item.icon}
                  </span>
                ) : null}
                <span>{item.label}</span>
              </Menu.Item>
            ))}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  );
}
