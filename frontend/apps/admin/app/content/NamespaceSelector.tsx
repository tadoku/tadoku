import { useEffect, useState } from 'react'
import { useRouter } from 'next/router'
import { Modal } from 'ui'

const STORAGE_KEY_NAMESPACES = 'admin_namespaces'
export const DEFAULT_NAMESPACE = 'tadoku'
const DEFAULT_NAMESPACES = [DEFAULT_NAMESPACE]
const NAMESPACE_PATTERN = /^[a-z0-9-]+$/

function loadNamespaces(): string[] {
  if (typeof window === 'undefined') return DEFAULT_NAMESPACES
  try {
    const stored = localStorage.getItem(STORAGE_KEY_NAMESPACES)
    if (stored) {
      const parsed = JSON.parse(stored)
      if (Array.isArray(parsed) && parsed.length > 0) return parsed
    }
  } catch {}
  return DEFAULT_NAMESPACES
}

export function useNamespace(): string {
  const router = useRouter()
  const ns = router.query.namespace
  if (typeof ns === 'string' && ns && NAMESPACE_PATTERN.test(ns)) return ns
  return DEFAULT_NAMESPACE
}

interface Props {
  value: string
  onChange: (namespace: string) => void
}

export function NamespaceSelector({ value, onChange }: Props) {
  const [namespaces, setNamespaces] = useState<string[]>(DEFAULT_NAMESPACES)
  const [adding, setAdding] = useState(false)
  const [newNamespace, setNewNamespace] = useState('')

  useEffect(() => {
    setNamespaces(loadNamespaces())
  }, [])

  const handleChange = (ns: string) => {
    onChange(ns)
  }

  const handleAdd = () => {
    const trimmed = newNamespace.trim().toLowerCase()
    if (!trimmed || !NAMESPACE_PATTERN.test(trimmed) || namespaces.includes(trimmed)) {
      setAdding(false)
      setNewNamespace('')
      return
    }
    const updated = [...namespaces, trimmed]
    setNamespaces(updated)
    try {
      localStorage.setItem(STORAGE_KEY_NAMESPACES, JSON.stringify(updated))
    } catch {}
    handleChange(trimmed)
    setAdding(false)
    setNewNamespace('')
  }

  const ADD_NEW_VALUE = '__add_new__'

  const handleSelectChange = (val: string) => {
    if (val === ADD_NEW_VALUE) {
      setAdding(true)
      return
    }
    handleChange(val)
  }

  return (
    <>
      <select
        className="input h-10 md:h-11 w-auto min-w-[140px] text-sm md:text-base"
        value={value}
        onChange={e => handleSelectChange(e.target.value)}
      >
        {namespaces.map(ns => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
        <option disabled>──────────</option>
        <option value={ADD_NEW_VALUE}>Add new...</option>
      </select>
      <Modal
        isOpen={adding}
        setIsOpen={setAdding}
        title="Add Namespace"
      >
        <form
          onSubmit={e => {
            e.preventDefault()
            handleAdd()
          }}
        >
          <label className="label">
            <span className="label-text">Namespace</span>
            <input
              type="text"
              className="input"
              placeholder="my-namespace"
              value={newNamespace}
              onChange={e => setNewNamespace(e.target.value)}
              autoFocus
            />
          </label>
          <div className="modal-actions justify-end">
            <button
              type="button"
              className="btn ghost"
              onClick={() => {
                setAdding(false)
                setNewNamespace('')
              }}
            >
              Cancel
            </button>
            <button type="submit" className="btn primary">
              Add
            </button>
          </div>
        </form>
      </Modal>
    </>
  )
}
