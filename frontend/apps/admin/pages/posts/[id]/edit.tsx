import { routes } from '@app/common/routes'
import { DocumentTextIcon, HomeIcon, PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Modal } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { ContentEditor } from '@app/content/ContentEditor'
import { postsConfig } from '@app/content/posts'
import { useRouter } from 'next/router'
import { useState } from 'react'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const id = router.query.id as string
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

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
          Are you sure you want to delete this post? This action is not yet
          supported by the API. Please contact a developer to delete content directly.
        </p>
        <div className="modal-actions">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setDeleteModalOpen(false)}
          >
            Close
          </button>
        </div>
      </Modal>
    </>
  )
}

Page.getLayout = getDashboardLayout('posts')

export default Page
