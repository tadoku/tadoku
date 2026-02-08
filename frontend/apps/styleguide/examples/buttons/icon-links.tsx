import {
  PencilSquareIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  ChevronLeftIcon,
  UserIcon,
} from '@heroicons/react/24/solid'

export default function IconButtonLinks() {
  return (
    <div className="space-x-3">
      <a href="#" className="btn primary">
        <PencilSquareIcon />
        Primary
      </a>
      <a href="#" className="btn secondary">
        <UserIcon />
        Secondary
      </a>
      <a href="#" className="btn">
        <ArrowDownTrayIcon />
        Tertiary
      </a>
      <a href="#" className="btn danger">
        <TrashIcon />
        Danger
      </a>
      <a href="#" className="btn ghost">
        <ChevronLeftIcon />
        Ghost
      </a>
    </div>
  )
}
