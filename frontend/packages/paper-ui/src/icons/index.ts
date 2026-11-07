import { cn } from "../foundations/variants";

// Outline icons are the Paper default for actions and navigation.
export {
  ArrowDownTrayIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ArrowRightStartOnRectangleIcon,
  Bars3Icon,
  BookOpenIcon,
  CalendarDaysIcon,
  ChatBubbleLeftRightIcon,
  Cog6ToothIcon,
  LanguageIcon,
  MoonIcon,
  ShieldCheckIcon,
  SunIcon,
  TrophyIcon,
  UserIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  EllipsisHorizontalIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  PlusIcon,
  TrashIcon,
  UserCircleIcon,
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
  compact: "paper-icon-compact inline-block flex-none align-middle size-icon-sm",
  default: "paper-icon-default inline-block flex-none align-middle size-icon",
  prominent: "paper-icon-prominent inline-block flex-none align-middle size-icon-lg",
  emptyState: "paper-icon-empty-state inline-block flex-none align-middle size-icon-xl",
} as const;

export type IconSize = keyof typeof iconSizeClassNames;

export function iconClassName(size: IconSize = "default", className?: string) {
  return cn(iconSizeClassNames[size], className);
}
