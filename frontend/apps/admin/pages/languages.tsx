import { routes } from '@app/common/routes'
import {
  HomeIcon,
  LanguageIcon,
  PencilSquareIcon,
} from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Loading, Modal } from 'ui'
import { NextPageWithLayout } from './_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import {
  useLanguageList,
  useLanguageCreate,
  useLanguageUpdate,
  Language,
} from '@app/languages/api'
import { iso639_3 } from '@app/languages/iso639-3'
import { Dispatch, SetStateAction, useState, useMemo, useEffect } from 'react'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

function WikipediaLink({ code }: { code: string }) {
  return (
    <a
      href={`https://en.wikipedia.org/wiki/ISO_639:${code}`}
      target="_blank"
      rel="noopener noreferrer"
      className="text-blue-600 hover:text-blue-800 underline"
    >
      {code}
    </a>
  )
}

interface LanguageFormProps {
  isOpen: boolean
  setIsOpen: Dispatch<SetStateAction<boolean>>
  editingLanguage: Language | null
  existingCodes: Set<string>
}

function LanguageForm({
  isOpen,
  setIsOpen,
  editingLanguage,
  existingCodes,
}: LanguageFormProps) {
  const [code, setCode] = useState('')
  const [name, setName] = useState('')
  const [search, setSearch] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const queryClient = useQueryClient()
  const isEditing = editingLanguage !== null

  const resetForm = () => {
    setCode('')
    setName('')
    setSearch('')
    setErrors({})
  }

  const createMutation = useLanguageCreate(
    () => {
      toast.success(`Language "${name}" created`)
      queryClient.invalidateQueries(['languages', 'list'])
      setIsOpen(false)
      resetForm()
    },
    error => {
      toast.error(error.message || 'Failed to create language')
    },
  )

  const updateMutation = useLanguageUpdate(
    () => {
      toast.success(`Language "${name}" updated`)
      queryClient.invalidateQueries(['languages', 'list'])
      setIsOpen(false)
      resetForm()
    },
    () => {
      toast.error('Failed to update language')
    },
  )

  // Sync form fields when editingLanguage changes
  useEffect(() => {
    if (isOpen && editingLanguage) {
      setCode(editingLanguage.code)
      setName(editingLanguage.name)
      setSearch('')
      setErrors({})
    } else if (isOpen && !editingLanguage) {
      resetForm()
    }
  }, [isOpen, editingLanguage])

  // Filter ISO 639-3 suggestions based on search input
  const suggestions = useMemo(() => {
    if (isEditing || !search.trim()) return []
    const query = search.toLowerCase()
    return iso639_3
      .filter(
        lang =>
          (lang.name.toLowerCase().includes(query) ||
            lang.code.toLowerCase().includes(query)) &&
          !existingCodes.has(lang.code),
      )
      .slice(0, 8)
  }, [search, isEditing, existingCodes])

  const handleSelectSuggestion = (lang: { code: string; name: string }) => {
    setCode(lang.code)
    setName(lang.name)
    setSearch('')
    setErrors({})
  }

  const handleClose = () => {
    setIsOpen(false)
    resetForm()
  }

  const handleSave = () => {
    const newErrors: Record<string, string> = {}
    if (!isEditing && !code.trim()) newErrors.code = 'Code is required'
    if (!isEditing && code.length > 10) newErrors.code = 'Code must be 10 characters or less'
    if (!name.trim()) newErrors.name = 'Name is required'
    if (name.length > 100) newErrors.name = 'Name must be 100 characters or less'

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }
    setErrors({})

    if (isEditing) {
      updateMutation.mutate({ code: editingLanguage.code, name: name.trim() })
    } else {
      createMutation.mutate({ code: code.trim(), name: name.trim() })
    }
  }

  const isSaving = createMutation.isLoading || updateMutation.isLoading

  return (
    <Modal
      isOpen={isOpen}
      setIsOpen={setIsOpen}
      title={isEditing ? 'Edit Language' : 'Add Language'}
    >
      <div className="modal-body flex flex-col gap-4">
        {!isEditing && (
          <div className="relative">
            <label className="label">
              <span className="label-text">
                Search ISO 639-3 languages
              </span>
              <input
                type="text"
                className="input"
                value={search}
                onChange={e => setSearch(e.target.value)}
                placeholder="Type a language name or code..."
              />
            </label>
            {suggestions.length > 0 && (
              <ul className="absolute z-10 left-0 right-0 bg-white border border-slate-200 rounded-md shadow-lg mt-1 max-h-48 overflow-y-auto">
                {suggestions.map(lang => (
                  <li key={lang.code}>
                    <button
                      type="button"
                      className="w-full text-left px-3 py-2 hover:bg-slate-100 flex justify-between items-center"
                      onClick={() => handleSelectSuggestion(lang)}
                    >
                      <span>{lang.name}</span>
                      <span className="text-slate-400 text-sm font-mono">
                        {lang.code}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        <label className={`label ${errors.code ? 'error' : ''}`}>
          <span className="label-text">Code (ISO 639-3)</span>
          <input
            type="text"
            className="input font-mono"
            value={code}
            onChange={e => {
              setCode(e.target.value)
              setErrors(prev => ({ ...prev, code: '' }))
            }}
            placeholder="e.g. jpn"
            disabled={isEditing}
            maxLength={10}
          />
          {errors.code && <span className="error">{errors.code}</span>}
          {code && !isEditing && (
            <span className="text-xs text-slate-500 mt-1">
              Wikipedia:{' '}
              <a
                href={`https://en.wikipedia.org/wiki/ISO_639:${code}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:text-blue-800 underline"
              >
                ISO 639:{code}
              </a>
            </span>
          )}
        </label>

        <label className={`label ${errors.name ? 'error' : ''}`}>
          <span className="label-text">Display Name</span>
          <input
            type="text"
            className="input"
            value={name}
            onChange={e => {
              setName(e.target.value)
              setErrors(prev => ({ ...prev, name: '' }))
            }}
            placeholder="e.g. Japanese"
            maxLength={100}
          />
          {errors.name && <span className="error">{errors.name}</span>}
        </label>
      </div>

      <div className="modal-actions">
        <button
          type="button"
          className="btn primary"
          onClick={handleSave}
          disabled={isSaving}
        >
          {isSaving
            ? 'Saving...'
            : isEditing
              ? 'Update Language'
              : 'Add Language'}
        </button>
        <button
          type="button"
          className="btn ghost"
          onClick={handleClose}
          disabled={isSaving}
        >
          Cancel
        </button>
      </div>
    </Modal>
  )
}

const Page: NextPageWithLayout = () => {
  const [modalOpen, setModalOpen] = useState(false)
  const [editingLanguage, setEditingLanguage] = useState<Language | null>(null)
  const [filter, setFilter] = useState('')

  const languages = useLanguageList({ enabled: true })

  const existingCodes = useMemo(
    () => new Set(languages.data?.languages.map(l => l.code) ?? []),
    [languages.data],
  )

  const filteredLanguages = useMemo(() => {
    if (!languages.data) return []
    if (!filter.trim()) return languages.data.languages
    const query = filter.toLowerCase()
    return languages.data.languages.filter(
      l =>
        l.name.toLowerCase().includes(query) ||
        l.code.toLowerCase().includes(query),
    )
  }, [languages.data, filter])

  const handleAdd = () => {
    setEditingLanguage(null)
    setModalOpen(true)
  }

  const handleEdit = (lang: Language) => {
    setEditingLanguage(lang)
    setModalOpen(true)
  }

  return (
    <>
      <Head>
        <title>Languages - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            {
              label: 'Admin',
              href: routes.home(),
              IconComponent: HomeIcon,
            },
            {
              label: 'Languages',
              href: routes.languages(),
              IconComponent: LanguageIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between">
        <h1 className="title">Languages</h1>
        <button type="button" className="btn primary" onClick={handleAdd}>
          Add Language
        </button>
      </div>

      <div className="mt-4">
        <input
          type="text"
          className="input"
          placeholder="Filter by name or code..."
          value={filter}
          onChange={e => setFilter(e.target.value)}
        />
      </div>

      {languages.isError ? (
        <div className="mt-4">
          {languages.error instanceof Error &&
          languages.error.message === '403' ? (
            <span className="flash error">
              You do not have permission to view this page.
            </span>
          ) : languages.error instanceof Error &&
            languages.error.message === '401' ? (
            <span className="flash error">
              You must be logged in to view this page.
            </span>
          ) : (
            <span className="flash error">Could not load languages.</span>
          )}
        </div>
      ) : null}

      {languages.isLoading ? (
        <div className="mt-4">
          <Loading />
        </div>
      ) : null}

      {languages.isSuccess ? (
        <div className="mt-4">
          <p className="text-sm text-slate-500 mb-2">
            {filteredLanguages.length} language
            {filteredLanguages.length !== 1 ? 's' : ''}
            {filter ? ' matching filter' : ' total'}
          </p>
          <div className="table-container">
            <table className="default">
              <thead>
                <tr>
                  <th className="default w-28">Code</th>
                  <th className="default">Name</th>
                  <th className="default w-32">Wikipedia</th>
                  <th className="default w-12"></th>
                </tr>
              </thead>
              <tbody>
                {filteredLanguages.map(lang => (
                  <tr key={lang.code}>
                    <td className="default font-mono">{lang.code}</td>
                    <td className="default font-medium">{lang.name}</td>
                    <td className="default">
                      <WikipediaLink code={lang.code} />
                    </td>
                    <td className="default">
                      <button
                        type="button"
                        className="btn ghost p-1"
                        onClick={() => handleEdit(lang)}
                        title="Edit language"
                      >
                        <PencilSquareIcon className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
                {filteredLanguages.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      {filter
                        ? 'No languages matching your filter'
                        : 'No languages found'}
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </div>
      ) : null}

      <LanguageForm
        isOpen={modalOpen}
        setIsOpen={setModalOpen}
        editingLanguage={editingLanguage}
        existingCodes={existingCodes}
      />
    </>
  )
}

Page.getLayout = getDashboardLayout('languages')

export default Page
