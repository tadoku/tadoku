import { routes } from '@app/common/routes'
import { DocumentDuplicateIcon, HomeIcon, PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Modal } from 'ui'
import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { pagesConfig } from '@app/content/pages'
import { useContentDelete } from '@app/content/api'
import { useNamespace } from '@app/content/NamespaceSelector'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const queryClient = useQueryClient()
  const id = router.query.id as string
  const namespace = useNamespace()
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  const deleteMutation = useContentDelete(
    pagesConfig,
    namespace,
    () => {
      toast.success('Page deleted successfully', { position: 'bottom-right' })
      queryClient.invalidateQueries([pagesConfig.type])
      router.push(routes.pages(namespace))
    },
    () => {
      toast.error('Failed to delete page', { position: 'bottom-right' })
    },
  )

  const handleDelete = () => {
    deleteMutation.mutate(id)
  }

  return (
    <>
      <Head>
        <title>Edit Page - Admin - Tadoku</title>
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
              label: 'Pages',
              href: routes.pages(namespace),
              IconComponent: DocumentDuplicateIcon,
            },
            {
              label: 'View',
              href: id ? routes.pagePreview(namespace, id) : '#',
            },
            {
              label: 'Edit',
              href: id ? routes.pageEdit(namespace, id) : '#',
              IconComponent: PencilIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Edit Page</h1>
        <button
          type="button"
          className="btn danger"
          onClick={() => setDeleteModalOpen(true)}
        >
          <TrashIcon className="w-4 h-4 mr-1 inline" />
          Delete
        </button>
      </div>
      {id ? <ContentEditor config={pagesConfig} id={id} /> : null}
      <Modal
        isOpen={deleteModalOpen}
        setIsOpen={setDeleteModalOpen}
        title="Delete Page"
      >
        <p className="modal-body">
          Are you sure you want to delete this page? This action cannot be
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

Page.getLayout = getDashboardLayout('pages')

export default Page
