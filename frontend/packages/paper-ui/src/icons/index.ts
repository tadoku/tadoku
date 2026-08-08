// Outline icons are the Paper default for actions and navigation.
export {
  ArrowLeftIcon,
  ArrowRightIcon,
  Bars3Icon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  EllipsisHorizontalIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  PlusIcon,
  TrashIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";

// Solid icons are reserved for status and confirmation.
export {
  CheckIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XCircleIcon,
} from "@heroicons/react/24/solid";

export const iconSizeClassNames = {
  compact: "paper-icon-compact",
  default: "paper-icon-default",
  prominent: "paper-icon-prominent",
  emptyState: "paper-icon-empty-state",
} as const;

export type IconSize = keyof typeof iconSizeClassNames;

export function iconClassName(size: IconSize = "default", className?: string) {
  return [iconSizeClassNames[size], className].filter(Boolean).join(" ");
}
