import { toast } from 'react-toastify'

export default function ToastExample() {
  return (
    <div className="h-stack spaced">
      <button className="btn" onClick={() => toast.info('Info toast')}>
        Info toast
      </button>
      <button className="btn" onClick={() => toast.success('Success toast')}>
        Success toast
      </button>
      <button className="btn" onClick={() => toast.warning('Warning toast')}>
        Warning toast
      </button>
      <button className="btn" onClick={() => toast.error('Error toast')}>
        Error toast
      </button>
    </div>
  )
}
