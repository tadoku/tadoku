import { routes } from '@app/common/routes'
import { HomeIcon, PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Modal } from 'ui'
import { ContentEditor } from './ContentEditor'
import { ContentConfig } from './types'
import { useContentDelete } from './api'
import { useNamespace } from './NamespaceSelector'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

interface Props {
  config: ContentConfig
}

export function ContentEditPage({ config }: Props) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const id = router.query.id as string
  const namespace = useNamespace()
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  const deleteMutation = useContentDelete(
    config,
    namespace,
    () => {
      toast.success(`${config.label} deleted successfully`, { position: 'bottom-right' })
      queryClient.invalidateQueries([config.type])
      router.push(config.routes.list(namespace))
    },
    () => {
      toast.error(`Failed to delete ${config.label.toLowerCase()}`, { position: 'bottom-right' })
    },
  )

  const handleDelete = () => {
    deleteMutation.mutate(id)
  }

  return (
    <>
      <Head>
        <title>Edit {config.label} - Admin - Tadoku</title>
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
              label: config.labelPlural,
              href: config.routes.list(namespace),
              IconComponent: config.icon,
            },
            {
              label: 'View',
              href: id ? config.routes.preview(namespace, id) : '#',
            },
            {
              label: 'Edit',
              href: id ? config.routes.edit(namespace, id) : '#',
              IconComponent: PencilIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Edit {config.label}</h1>
        <button
          type="button"
          className="btn danger"
          onClick={() => setDeleteModalOpen(true)}
        >
          <TrashIcon className="w-4 h-4 mr-1 inline" />
          Delete
        </button>
      </div>
      {id ? <ContentEditor config={config} id={id} /> : null}
      <Modal
        isOpen={deleteModalOpen}
        setIsOpen={setDeleteModalOpen}
        title={`Delete ${config.label}`}
      >
        <p className="modal-body">
          Are you sure you want to delete this {config.label.toLowerCase()}? This action cannot be
          undone.
        </p>
        <div className="modal-actions justify-end">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setDeleteModalOpen(false)}
            disabled={deleteMutation.isLoading}
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn danger"
            onClick={handleDelete}
            disabled={deleteMutation.isLoading}
          >
            {deleteMutation.isLoading ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </Modal>
    </>
  )
}
