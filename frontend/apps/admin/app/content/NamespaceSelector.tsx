import { useEffect, useState } from 'react'
import { PlusIcon } from '@heroicons/react/20/solid'

const STORAGE_KEY_NAMESPACES = 'admin_namespaces'
const STORAGE_KEY_SELECTED = 'admin_namespace'
const DEFAULT_NAMESPACES = ['tadoku']
const DEFAULT_SELECTED = 'tadoku'

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

function loadSelected(): string {
  if (typeof window === 'undefined') return DEFAULT_SELECTED
  try {
    const stored = localStorage.getItem(STORAGE_KEY_SELECTED)
    if (stored) return stored
  } catch {}
  return DEFAULT_SELECTED
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
    try {
      localStorage.setItem(STORAGE_KEY_SELECTED, ns)
    } catch {}
  }

  const handleAdd = () => {
    const trimmed = newNamespace.trim().toLowerCase()
    if (!trimmed || namespaces.includes(trimmed)) {
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

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-semibold text-slate-600">Namespace:</label>
      <select
        className="input h-9 w-auto min-w-[140px] text-sm"
        value={value}
        onChange={e => handleChange(e.target.value)}
      >
        {namespaces.map(ns => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
      </select>
      {adding ? (
        <form
          className="flex items-center gap-1"
          onSubmit={e => {
            e.preventDefault()
            handleAdd()
          }}
        >
          <input
            type="text"
            className="input h-9 w-32 text-sm"
            placeholder="namespace"
            value={newNamespace}
            onChange={e => setNewNamespace(e.target.value)}
            autoFocus
          />
          <button type="submit" className="btn small secondary">
            Add
          </button>
          <button
            type="button"
            className="btn small ghost"
            onClick={() => {
              setAdding(false)
              setNewNamespace('')
            }}
          >
            Cancel
          </button>
        </form>
      ) : (
        <button
          type="button"
          className="btn small ghost"
          onClick={() => setAdding(true)}
          title="Add namespace"
        >
          <PlusIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  )
}

export function useNamespace() {
  const [namespace, setNamespace] = useState(DEFAULT_SELECTED)

  useEffect(() => {
    setNamespace(loadSelected())
  }, [])

  const handleChange = (ns: string) => {
    setNamespace(ns)
    try {
      localStorage.setItem(STORAGE_KEY_SELECTED, ns)
    } catch {}
  }

  return [namespace, handleChange] as const
}
