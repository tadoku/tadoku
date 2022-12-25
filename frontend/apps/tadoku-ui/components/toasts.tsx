import 'react-toastify/dist/ReactToastify.css'
import { ToastContainer as Container } from 'react-toastify'

export default function ToastContainer() {
  return (
    <Container
      toastClassName={() =>
        'relative flex px-3 py-3 text-black border mb-1 border-b-2 shadow shadow-slate-500/10 bg-white border-black/10 justify-between overflow-hidden cursor-pointer'
      }
    />
  )
}
