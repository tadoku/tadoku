import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { Link, useLocation, useParams } from 'react-router-dom'
import { Breadcrumb, Button, Drawer, Flash, Input, Modal, Select, Sidebar, Table, TextArea, buttonClassName, type NavigationLinkProps, type SidebarSection } from 'paper-ui'
import { Bars3Icon, PlusIcon } from 'paper-ui/icons'
import { formatDate, users } from '../data'
import { usePlayground } from '../state'
import './records.css'

type AdminSection = 'posts' | 'pages' | 'announcements' | 'languages' | 'reports'
type AdminRecord = { id: string; section: AdminSection; title: string; detail: string; status: string }
const initialRecords: AdminRecord[] = [
  { id: 'post-reading', section: 'posts', title: 'Make room for a little reading', detail: 'A community note about choosing material that makes you want to come back tomorrow.', status: 'Published' },
  { id: 'post-round5', section: 'posts', title: 'Round 5 reading notes', detail: 'Collect recommendations and reflections from this month’s participants.', status: 'Draft' },
  { id: 'page-about', section: 'pages', title: 'About Tadoku', detail: 'A community built around spending time with the languages you want to understand.', status: 'Published' },
  { id: 'page-guide', section: 'pages', title: 'Start here', detail: 'Choose how you participate, read or listen, then record your activity.', status: 'Published' },
  { id: 'announcement-round5', section: 'announcements', title: 'September Round 5 is open', detail: 'Read and listen together from 1–30 September. Registration closes 23 September, UTC.', status: 'Draft' },
  { id: 'announcement-round6', section: 'announcements', title: 'The November round', detail: 'The next official round takes place from 1–14 November.', status: 'Draft' },
  { id: 'language-ja', section: 'languages', title: 'Japanese', detail: 'Language code: ja. Available for reading and listening.', status: 'Approved' },
  { id: 'language-fr', section: 'languages', title: 'French', detail: 'Language code: fr. Available for reading and listening.', status: 'Approved' },
  { id: 'language-ko', section: 'languages', title: 'Korean request', detail: 'A member requested Korean for their next immersion contest. Check the supported-language list before adding a duplicate.', status: 'Open' },
  { id: 'language-cy', section: 'languages', title: 'Welsh request', detail: 'A member requested Welsh. Review the name and language code before approval.', status: 'Open' },
  { id: 'report-title', section: 'reports', title: 'Unclear activity description', detail: 'A participant asked for clarification about a sample entry. Review its activity and scoring context before taking action.', status: 'Open' },
  { id: 'report-duplicate', section: 'reports', title: 'Possible duplicate submission', detail: 'Two sample records have similar descriptions. Compare their dates and source material; matching titles alone do not establish a duplicate.', status: 'Open' },
  { id: 'report-profile', section: 'reports', title: 'Profile name review', detail: 'A sample display name was flagged for moderator review. Record the review outcome here.', status: 'Resolved' },
]
const names: Record<string, string> = { dashboard: 'Dashboard', posts: 'Posts', pages: 'Pages', announcements: 'Announcements', languages: 'Languages', contests: 'Contests', users: 'Users', reports: 'Reports' }
const sections: SidebarSection[] = [
  { id: 'overview', title: 'Overview', links: [{ id: 'dashboard', label: 'Dashboard', href: '/admin' }] },
  { id: 'content', title: 'Content', links: ['posts', 'pages', 'announcements'].map(id => ({ id, label: names[id], href: `/admin/${id}` })) },
  { id: 'platform', title: 'Platform', links: ['languages', 'contests'].map(id => ({ id, label: names[id], href: `/admin/${id}` })) },
  { id: 'moderation', title: 'Moderation', links: ['users', 'reports'].map(id => ({ id, label: names[id], href: `/admin/${id}` })) },
]
const renderLink = ({ href, ...props }: NavigationLinkProps) => <Link className="text-link" to={href} {...props} />

function RecordEditor({ record, section, onSave, variant }: { record?: AdminRecord; section: AdminSection; variant?: 'default' | 'outline'; onSave: (record: AdminRecord) => void }) {
  const [open, setOpen] = useState(false)
  const review = section === 'reports' || section === 'languages'
  const defaults = { title: record?.title ?? '', detail: record?.detail ?? '', status: record?.status ?? (review ? 'Open' : 'Draft') }
  const methods = useForm({ defaultValues: defaults })
  const singular: Record<AdminSection, string> = { posts: 'post', pages: 'page', announcements: 'announcement', languages: 'language', reports: 'report' }
  const title = record ? `${review ? 'Review' : 'Edit'} ${singular[section]}` : `New ${singular[section]}`
  return <Modal title={title} trigger={<Button variant={variant ?? (record ? 'outline' : 'default')} leadingIcon={record ? undefined : <PlusIcon className="paper-icon-default" />}>{record ? review ? 'Review' : 'Edit' : title}</Button>} open={open} onOpenChange={next => { if (next) methods.reset(defaults); setOpen(next) }} footer={null}>
    <FormProvider {...methods}><form onSubmit={methods.handleSubmit(values => { onSave({ ...values, id: record?.id ?? `${section}-${crypto.randomUUID()}`, section }); setOpen(false) })}>
      <div className="app-form__fields">
      <Input name="title" label={section === 'languages' ? 'Language or request' : 'Title'} required maxLength={120} rules={{ validate: value => Boolean(String(value).trim()) || 'Enter a title.' }} />
      <TextArea name="detail" label={review ? 'Review notes' : 'Content'} required rows={6} rules={{ validate: value => Boolean(String(value).trim()) || 'Add the content or review notes.' }} />
      <Select name="status" label="Status" options={(review ? ['Open', section === 'languages' ? 'Approved' : 'Resolved'] : ['Draft', 'Published']).map(value => ({ value, label: value }))} />
      </div>
      <div className="app-form__actions"><Button type="submit">Save {review ? 'review' : 'changes'}</Button><Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button></div>
    </form></FormProvider>
  </Modal>
}

export function AdminPage() {
  const { section: requestedSection } = useParams()
  const section = requestedSection ?? 'dashboard'
  const location = useLocation()
  const { contests, logs, viewer, setViewer } = usePlayground()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [records, setRecords] = useState(initialRecords)
  const [notice, setNotice] = useState('')
  const currentContest = contests.find(contest => contest.id === 'round5')
  const openReports = records.filter(record => record.section === 'reports' && record.status === 'Open')
  const requestedLanguages = records.filter(record => record.section === 'languages' && record.status === 'Open')
  const saveRecord = (record: AdminRecord) => {
    setRecords(items => items.some(item => item.id === record.id) ? items.map(item => item.id === record.id ? record : item) : [record, ...items])
    setNotice(`${record.title}: ${record.status.toLocaleLowerCase()} saved in this workspace.`)
  }
  const sidebar = <Sidebar sections={sections} currentPath={location.pathname} label="Admin sections" renderLink={({ href, ...props }) => <Link className="text-link" to={href} {...props} onClick={() => setMobileOpen(false)} />} />

  if (viewer !== 'admin') return <section className="empty-state"><h1 className="paper-type-page">Admin workspace</h1><p>This workspace is available to administrators.</p><Button onClick={() => setViewer('admin')}>Enter the demo admin workspace</Button></section>

  return <div className="records-admin-layout">
    <aside className="records-admin-sidebar">{sidebar}<Link className="records-admin-return text-link" to="/">Back to Tadoku</Link></aside>
    <div className="records-admin-main">
      <div className="records-admin-mobile"><Drawer trigger={<Button variant="outline" leadingIcon={<Bars3Icon className="paper-icon-default" />}>Admin sections</Button>} title="Admin workspace" placement="start" open={mobileOpen} onOpenChange={setMobileOpen}>{sidebar}<Link className="text-link" to="/" onClick={() => setMobileOpen(false)}>Back to Tadoku</Link></Drawer></div>
      <Breadcrumb items={[{ id: 'home', label: 'Tadoku', href: '/' }, { id: 'admin', label: 'Admin', href: '/admin' }, { id: section, label: names[section] ?? 'Page not found' }]} renderLink={renderLink} />
      <header className="page-header records-admin-heading"><div><h1 className="paper-type-page">{names[section] ?? 'Page not found'}</h1></div>{section === 'dashboard' ? <RecordEditor section="announcements" onSave={saveRecord} /> : ['posts', 'pages', 'announcements', 'languages'].includes(section) ? <RecordEditor section={section as AdminSection} onSave={saveRecord} /> : null}</header>
      {notice ? <Flash variant="success" className="records-admin-notice">{notice}</Flash> : null}
      {section === 'dashboard' ? <>
        <div className="records-admin-stats"><div className="stat"><span>Community members</span><strong>{users.length}</strong></div><div className="stat"><span>Active official round</span><strong>Round 5</strong><small>1–30 September 2026</small></div><div className="stat"><span>Open reports</span><strong>{openReports.length}</strong></div></div>
        <section className="page-section records-admin-attention"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section">Needs attention</h2></div><Link className="text-link" to="/admin/reports">View reports</Link></div><ul className="records-admin-queue">
          {openReports.map(record => <li key={record.id}><div><strong>{record.title}</strong><p>{record.detail}</p></div><RecordEditor record={record} section="reports" onSave={saveRecord} /></li>)}
          {records.filter(record => record.section === 'announcements' && record.status === 'Draft').slice(0, 1).map(record => <li key={record.id}><div><strong>{record.title}</strong><p>Draft announcement ready for an editorial check.</p></div><RecordEditor record={record} section="announcements" onSave={saveRecord} /></li>)}
          {requestedLanguages.length ? <li><div><strong>{requestedLanguages.length} language requests</strong><p>Check names and supported-language entries before approval.</p></div><Link className={buttonClassName({ variant: 'outline' })} to="/admin/languages">Review requests</Link></li> : null}
          {!openReports.length && !requestedLanguages.length ? <li><p>All reports and language requests have been reviewed.</p></li> : null}
        </ul></section>
        <div className="records-admin-bottom">
          <section className="page-section"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section">Current contest</h2><p className="muted">{currentContest?.title ?? '2026 Round 5'}</p></div><Link className="text-link" to="/contests/round5">Open contest</Link></div><dl className="records-admin-contest"><div><dt>Period</dt><dd>1–30 September, UTC</dd></div><div><dt>Registration deadline</dt><dd>23 September, UTC</dd></div><div><dt>Submitted sample entries</dt><dd>{logs.filter(log => log.submissions.some(item => item.contestId === 'round5')).length}</dd></div></dl></section>
          <section className="page-section"><div className="section-heading records-section-heading"><div><h2 className="paper-type-section">Quick actions</h2></div></div><div className="records-admin-quick"><RecordEditor section="posts" variant="outline" onSave={saveRecord} /><RecordEditor section="pages" variant="outline" onSave={saveRecord} /><Link className={buttonClassName({ variant: 'outline' })} to="/admin/languages">Languages</Link><Link className={buttonClassName({ variant: 'outline' })} to="/admin/users">Users</Link></div></section>
        </div>
        <section className="page-section records-admin-status"><h2 className="paper-type-section">Service status</h2><p className="muted">Illustrative service states for this unpublished preview.</p><dl><div><dt>Immersion API</dt><dd>Healthy sample</dd></div><div><dt>Content API</dt><dd>Healthy sample</dd></div><div><dt>Background jobs</dt><dd>Healthy sample</dd></div></dl></section>
      </> : section === 'users' ? <Table className="records-admin-table" caption="Community members" captionVisibility="screen-reader" minWidth="32rem" columns={[{ id: 'name', header: 'Member', rowHeader: true, cell: item => <Link className="text-link" to={`/users/${item.id}`}>{item.name}</Link> }, { id: 'joined', header: 'Tracking since', cell: item => formatDate(item.joined) }, { id: 'action', header: 'Activity', cell: item => <Link className="text-link" to={`/users/${item.id}/activity`}>View activity</Link> }]} rows={users} getRowKey={item => item.id} /> : section === 'contests' ? <Table className="records-admin-table" caption="Contests" captionVisibility="screen-reader" minWidth="38rem" columns={[{ id: 'title', header: 'Contest', rowHeader: true, cell: item => <Link className="text-link" to={`/contests/${item.id}`}>{item.title}</Link> }, { id: 'type', header: 'Collection', cell: item => item.unlisted ? 'Unlisted' : item.scope === 'official' ? 'Official' : 'Community' }, { id: 'dates', header: 'Begins', cell: item => formatDate(item.start) }, { id: 'action', header: 'Details', cell: item => <Link className="text-link" to={`/contests/${item.id}`}>Open contest</Link> }]} rows={contests} getRowKey={item => item.id} /> : ['posts', 'pages', 'announcements', 'languages', 'reports'].includes(section) ? <section className="records-admin-collection" aria-label={names[section]}>
        <p className="records-admin-count">{records.filter(record => record.section === section).length} {names[section].toLocaleLowerCase()}</p>
        <ul className="records-admin-records">{records.filter(record => record.section === section).map(item => <li key={item.id}>
          <div className="records-admin-record"><h2>{item.title}</h2><p>{item.detail}</p></div>
          <span className="status-tag">{item.status}</span>
          <RecordEditor record={item} section={item.section} onSave={saveRecord} />
        </li>)}</ul>
        {!records.some(record => record.section === section) ? <p className="empty-state">No {names[section].toLocaleLowerCase()} yet.</p> : null}
      </section> : <section className="empty-state"><h2 className="paper-type-section">This admin page isn’t available</h2><Link className={buttonClassName()} to="/admin">Return to dashboard</Link></section>}
    </div>
  </div>
}
