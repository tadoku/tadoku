import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon, PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Modal } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { postsConfig } from '@app/content/posts'
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
  const [namespace] = useNamespace()
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  const deleteMutation = useContentDelete(
    postsConfig,
    namespace,
    () => {
      toast.success('Post deleted successfully')
      queryClient.invalidateQueries([postsConfig.type])
      router.push(routes.posts())
    },
    () => {
      toast.error('Failed to delete post')
    },
  )

  const handleDelete = () => {
    deleteMutation.mutate(id)
  }

  return (
    <>
      <Head>
        <title>Edit Post - Admin - Tadoku</title>
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
              label: 'Posts',
              href: routes.posts(),
              IconComponent: DocumentTextIcon,
            },
            {
              label: 'View',
              href: id ? routes.postPreview(id) : '#',
            },
            {
              label: 'Edit',
              href: id ? routes.postEdit(id) : '#',
              IconComponent: PencilIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Edit Post</h1>
        <button
          type="button"
          className="btn danger"
          onClick={() => setDeleteModalOpen(true)}
        >
          <TrashIcon className="w-4 h-4 mr-1 inline" />
          Delete
        </button>
      </div>
      {id ? <ContentEditor config={postsConfig} id={id} /> : null}
      <Modal
        isOpen={deleteModalOpen}
        setIsOpen={setDeleteModalOpen}
        title="Delete Post"
      >
        <p className="modal-body">
          Are you sure you want to delete this post? This action cannot be
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

Page.getLayout = getDashboardLayout('posts')

export default Page
