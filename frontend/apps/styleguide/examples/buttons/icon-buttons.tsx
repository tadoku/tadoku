import {
  PencilSquareIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  ChevronLeftIcon,
  UserIcon,
} from '@heroicons/react/24/solid'

export default function IconButtons() {
  return (
    <div className="h-stack spaced">
      <button className="btn primary">
        <PencilSquareIcon />
        Primary
      </button>
      <button className="btn secondary">
        <UserIcon />
        Secondary
      </button>
      <button className="btn">
        <ArrowDownTrayIcon />
        Tertiary
      </button>
      <button className="btn danger">
        <TrashIcon />
        Danger
      </button>
      <button className="btn ghost">
        <ChevronLeftIcon />
        Ghost
      </button>
    </div>
  )
}
