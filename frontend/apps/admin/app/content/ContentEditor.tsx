import { useEffect, useState } from 'react'
import { FormProvider, useForm, useController } from 'react-hook-form'
import { Input, Loading, Tabbar } from 'ui'
import { ContentConfig } from './types'
import { useContentCreate, useContentFindById, useContentUpdate } from './api'
import { useNamespace } from './NamespaceSelector'
import { useRouter } from 'next/router'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'
import { v4 as uuidv4 } from 'uuid'
import { toUtcISOStringFromLocal } from '@app/common/datetime'

interface Props {
  config: ContentConfig
  id?: string // undefined = new item, string = editing existing
}

interface ContentFormValues {
  title: string
  slug: string
  body: string
  publishedAt: string
}

export function ContentEditor({ config, id }: Props) {
  const namespace = useNamespace()
  const router = useRouter()
  const queryClient = useQueryClient()
  const isNew = !id

  // Mobile tab state
  const [mobileTab, setMobileTab] = useState<'edit' | 'preview'>('edit')

  // Form state
  const methods = useForm<ContentFormValues>({
    defaultValues: {
      title: '',
      slug: '',
      body: '',
      publishedAt: '',
    },
  })

  const [initialized, setInitialized] = useState(false)

  // Load existing item if editing
  const existing = useContentFindById(config, namespace, id ?? '', {
    enabled: !isNew && !!id && !!namespace,
  })

  const slugify = (value: string) =>
    value
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')

  // Populate form when data loads
  useEffect(() => {
    if (existing.data && !initialized) {
      let formattedBody: string
      try {
        formattedBody = config.formatBody(existing.data.body)
      } catch {
        formattedBody = existing.data.body
      }
      methods.reset({
        title: existing.data.title,
        slug: existing.data.slug,
        body: formattedBody,
        publishedAt: existing.data.published_at
          ? existing.data.published_at.slice(0, 16) // format for datetime-local input
          : '',
      })
      setInitialized(true)
    }
  }, [existing.data, initialized, methods, config])

  // Auto-generate slug from title for new items
  const titleValue = methods.watch('title')
  useEffect(() => {
    if (isNew) {
      methods.setValue('slug', slugify(titleValue))
    }
  }, [titleValue, isNew, methods])

  // Body controller for custom editor
  const bodyController = useController({
    name: 'body',
    control: methods.control,
    rules: { validate: v => v.trim() !== '' || 'Content is required' },
  })

  // Watch values for live preview
  const watchedTitle = methods.watch('title')
  const watchedBody = methods.watch('body')

  const createMutation = useContentCreate(
    config,
    namespace,
    () => {
      toast.success(`${config.label} created successfully`, { position: 'bottom-right' })
      queryClient.removeQueries([config.type])
    },
    () => {
      toast.error(`Failed to create ${config.label.toLowerCase()}`, { position: 'bottom-right' })
    },
  )

  const updateMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success(`${config.label} updated successfully`, { position: 'bottom-right' })
      queryClient.removeQueries([config.type])
    },
    () => {
      toast.error(`Failed to update ${config.label.toLowerCase()}`, { position: 'bottom-right' })
    },
  )

  const isSaving = createMutation.isLoading || updateMutation.isLoading

  const handleSave = methods.handleSubmit(data => {
    const itemId = isNew ? uuidv4() : existing.data!.id
    const input = {
      id: itemId,
      slug: data.slug.trim().toLowerCase(),
      title: data.title.trim(),
      body: data.body,
      published_at: data.publishedAt ? toUtcISOStringFromLocal(data.publishedAt) : null,
    }

    const onSuccess = () => router.push(config.routes.preview(namespace, itemId))

    if (isNew) {
      createMutation.mutate(input, { onSuccess })
    } else {
      updateMutation.mutate(input, { onSuccess })
    }
  })

  if (!isNew && existing.isLoading) {
    return <Loading />
  }

  if (!isNew && existing.isError) {
    return (
      <span className="flash error">
        {existing.error instanceof Error && existing.error.message === '404'
          ? `${config.label} not found.`
          : `Could not load ${config.label.toLowerCase()}.`}
      </span>
    )
  }

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col gap-6">
        {/* Mobile tab bar */}
        <div className="lg:hidden">
          <Tabbar
            alwaysExpanded
            links={[
              { label: 'Edit', active: mobileTab === 'edit', onClick: () => setMobileTab('edit') },
              { label: 'Preview', active: mobileTab === 'preview', onClick: () => setMobileTab('preview') },
            ]}
          />
        </div>

        <div className="flex flex-col lg:flex-row gap-6">
          {/* Edit form */}
          <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'preview' ? 'hidden lg:flex' : 'flex'}`}>
            <div className="card flex flex-col gap-4">
              <Input
                name="title"
                type="text"
                label="Title"
                placeholder={`${config.label} title`}
                options={{ required: 'Title is required' }}
              />
              <div className={`label ${methods.formState.errors.slug ? 'error' : ''}`}>
                <div className="flex items-center justify-between">
                  <span className="label-text">Slug</span>
                  <button
                    type="button"
                    className="btn ghost"
                    onClick={() => methods.setValue('slug', slugify(methods.getValues('title')))}
                    disabled={!methods.getValues('title')?.trim()}
                  >
                    Regenerate
                  </button>
                </div>
                <input
                  type="text"
                  className="input"
                  placeholder="url-friendly-slug"
                  {...methods.register('slug', { validate: v => v.trim() !== '' || 'Slug is required' })}
                />
                <span className="error">{methods.formState.errors.slug?.message}</span>
              </div>
              <div className={`label flex-1 ${methods.formState.errors.body ? 'error' : ''}`}>
                <div className="flex items-center justify-between">
                  <span className="label-text">Content</span>
                  <button
                    type="button"
                    className="btn ghost"
                    onClick={() => {
                      try {
                        bodyController.field.onChange(config.formatBody(bodyController.field.value))
                      } catch (e) {
                        toast.error(
                          `Format failed: ${e instanceof Error ? e.message : 'unknown error'}`,
                          { position: 'bottom-right' },
                        )
                      }
                    }}
                  >
                    Format
                  </button>
                </div>
                {config.renderEditor({
                  value: bodyController.field.value,
                  onChange: bodyController.field.onChange,
                  placeholder: `Write your ${config.label.toLowerCase()} content here...`,
                })}
                <span className="error">{methods.formState.errors.body?.message}</span>
              </div>
              <Input
                name="publishedAt"
                type="datetime-local"
                label="Publish Date (UTC)"
                hint="Leave empty to save as draft"
              />
              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  className="btn ghost"
                  onClick={() => router.back()}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="btn primary"
                  onClick={handleSave}
                  disabled={isSaving}
                >
                  {isSaving ? 'Saving...' : isNew ? `Create ${config.label}` : `Save ${config.label}`}
                </button>
              </div>
            </div>
          </div>

          {/* Live preview */}
          <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'edit' ? 'hidden lg:flex' : 'flex'}`}>
            <div className="card flex-1 overflow-auto" style={{ minHeight: '500px' }}>
              {watchedTitle.trim() || watchedBody.trim() ? (
                <>
                  {watchedTitle.trim() ? <h2 className="text-xl font-bold mb-4">{watchedTitle}</h2> : null}
                  {config.renderBody(watchedBody)}
                </>
              ) : (
                <p className="text-sm text-slate-400 italic">
                  Start typing to see a preview...
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </FormProvider>
  )
}
