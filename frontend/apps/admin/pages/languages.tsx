import { routes } from '@app/common/routes'
import {
  HomeIcon,
  LanguageIcon,
  PencilSquareIcon,
} from '@heroicons/react/20/solid'
import Head from 'next/head'
import { AutocompleteInput, Breadcrumb, Input, Loading, Modal } from 'ui'
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
import { FormProvider, useForm } from 'react-hook-form'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

function WikipediaLink({ code }: { code: string }) {
  return (
    <a
      href={`https://en.wikipedia.org/wiki/ISO_639:${code}`}
      target="_blank"
      rel="noopener noreferrer"
      className="tag bg-slate-100 text-slate-600 font-mono hover:bg-slate-200 inline-flex"
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
  const queryClient = useQueryClient()
  const isEditing = editingLanguage !== null

  const methods = useForm({
    defaultValues: {
      isoLanguage: null as { code: string; name: string } | null,
      code: '',
      name: '',
    },
  })

  const createMutation = useLanguageCreate(
    () => {
      toast.success(`Language "${methods.getValues('name')}" created`)
      queryClient.invalidateQueries(['languages', 'list'])
      setIsOpen(false)
      methods.reset()
    },
    error => {
      toast.error(error.message || 'Failed to create language')
    },
  )

  const updateMutation = useLanguageUpdate(
    () => {
      toast.success(`Language "${methods.getValues('name')}" updated`)
      queryClient.invalidateQueries(['languages', 'list'])
      setIsOpen(false)
      methods.reset()
    },
    () => {
      toast.error('Failed to update language')
    },
  )

  // Sync form when modal opens
  useEffect(() => {
    if (isOpen && editingLanguage) {
      methods.reset({
        isoLanguage: null,
        code: editingLanguage.code,
        name: editingLanguage.name,
      })
    } else if (isOpen && !editingLanguage) {
      methods.reset({ isoLanguage: null, code: '', name: '' })
    }
  }, [isOpen, editingLanguage, methods])

  // When ISO language is selected from autocomplete, populate code and name
  const isoLanguage = methods.watch('isoLanguage')
  useEffect(() => {
    if (isoLanguage) {
      methods.setValue('code', isoLanguage.code)
      methods.setValue('name', isoLanguage.name)
    }
  }, [isoLanguage, methods])

  const availableLanguages = useMemo(
    () => iso639_3.filter(lang => !existingCodes.has(lang.code)),
    [existingCodes],
  )

  const isoCodes = useMemo(
    () => new Set(iso639_3.map(lang => lang.code)),
    [],
  )

  const codeValue = methods.watch('code')
  const codeNotInISO = !isEditing && codeValue.trim() !== '' && !isoCodes.has(codeValue.trim())

  const handleSave = methods.handleSubmit(data => {
    if (isEditing) {
      updateMutation.mutate({ code: editingLanguage!.code, name: data.name.trim() })
    } else {
      createMutation.mutate({ code: data.code.trim(), name: data.name.trim() })
    }
  })

  const isSaving = createMutation.isLoading || updateMutation.isLoading

  return (
    <Modal
      isOpen={isOpen}
      setIsOpen={setIsOpen}
      title={isEditing ? 'Edit Language' : 'Add Language'}
    >
      <FormProvider {...methods}>
        <div className="modal-body flex flex-col gap-4">
          {!isEditing && (
            <AutocompleteInput
              name="isoLanguage"
              label="Search ISO 639-3 languages"
              options={availableLanguages}
              match={(option, query) =>
                option.name.toLowerCase().includes(query.toLowerCase()) ||
                option.code.toLowerCase().includes(query.toLowerCase())
              }
              format={option => `${option.name} (${option.code})`}
            />
          )}

          <div>
            <Input
              name="code"
              type="text"
              label="Code (ISO 639-3)"
              placeholder="e.g. jpn"
              disabled={isEditing}
              maxLength={10}
              className="font-mono"
              options={
                !isEditing
                  ? {
                      required: 'Code is required',
                      maxLength: {
                        value: 10,
                        message: 'Code must be 10 characters or less',
                      },
                    }
                  : undefined
              }
            />
            {codeNotInISO && (
              <span className="text-xs text-amber-600 mt-1 block">
                This code is not in the ISO 639-3 list. Make sure this is intentional.
              </span>
            )}
            {codeValue && !isEditing && (
              <span className="text-xs text-slate-500 mt-1 block">
                Wikipedia:{' '}
                <a
                  href={`https://en.wikipedia.org/wiki/ISO_639:${codeValue}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-800 underline"
                >
                  ISO 639:{codeValue}
                </a>
              </span>
            )}
          </div>

          <Input
            name="name"
            type="text"
            label="Display Name"
            placeholder="e.g. Japanese"
            maxLength={100}
            options={{
              required: 'Name is required',
              maxLength: {
                value: 100,
                message: 'Name must be 100 characters or less',
              },
            }}
          />
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
            onClick={() => {
              setIsOpen(false)
              methods.reset()
            }}
            disabled={isSaving}
          >
            Cancel
          </button>
        </div>
      </FormProvider>
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
                  <th className="default w-12"></th>
                </tr>
              </thead>
              <tbody>
                {filteredLanguages.map(lang => (
                  <tr key={lang.code}>
                    <td className="default">
                      <WikipediaLink code={lang.code} />
                    </td>
                    <td className="default font-medium">{lang.name}</td>
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
                      colSpan={3}
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
