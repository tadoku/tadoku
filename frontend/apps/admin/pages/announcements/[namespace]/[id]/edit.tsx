import { routes } from '@app/common/routes'
import { MegaphoneIcon, HomeIcon, PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Modal } from 'ui'
import { NextPageWithLayout } from '../../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { AnnouncementEditor } from '@app/announcements/AnnouncementEditor'
import { useAnnouncementDelete, useAnnouncementFind } from '@app/announcements/api'
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
  const item = useAnnouncementFind(namespace, id, { enabled: !!id })
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  const deleteMutation = useAnnouncementDelete(
    namespace,
    () => {
      toast.success('Announcement deleted successfully', { position: 'bottom-right' })
      queryClient.invalidateQueries(['announcements'])
      router.push(routes.announcements(namespace))
    },
    () => {
      toast.error('Failed to delete announcement', { position: 'bottom-right' })
    },
  )

  const handleDelete = () => {
    deleteMutation.mutate(id)
  }

  return (
    <>
      <Head>
        <title>Edit Announcement - Admin - Tadoku</title>
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
              label: 'Announcements',
              href: routes.announcements(namespace),
              IconComponent: MegaphoneIcon,
            },
            {
              label: item.data?.title ?? 'Edit',
              href: id ? routes.announcementEdit(namespace, id) : '#',
              IconComponent: PencilIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Edit Announcement</h1>
        <button
          type="button"
          className="btn danger"
          onClick={() => setDeleteModalOpen(true)}
        >
          <TrashIcon className="w-4 h-4 mr-1 inline" />
          Delete
        </button>
      </div>
      {id ? <AnnouncementEditor id={id} /> : null}
      <Modal
        isOpen={deleteModalOpen}
        setIsOpen={setDeleteModalOpen}
        title="Delete Announcement"
      >
        <p className="modal-body">
          Are you sure you want to delete this announcement? This action cannot be undone.
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

Page.getLayout = getDashboardLayout('announcements')

export default Page
