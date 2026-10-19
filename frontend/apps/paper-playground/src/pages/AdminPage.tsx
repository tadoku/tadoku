import { useState } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { usePaperForm, Breadcrumb, Button, DescriptionList, Drawer, EmptyState, Flash, Modal, PageHeader, SectionHeading, Sidebar, StatGrid, StatusTag, Table, buttonClassName, type SidebarSection, type StatusTagTone } from 'paper-ui'
import { Bars3Icon, PlusIcon } from 'paper-ui/icons'
import { formatDate, users } from '../data'
import { contestEnd, contestStart, formatDateRange, formatDateTime } from '../dates'
import { renderLink } from '../routerLink'
import { usePlayground } from '../state'
import './records.css'

type AdminSection = 'posts' | 'pages' | 'announcements' | 'languages' | 'reports'
type AdminRecord = { id: string; section: AdminSection; title: string; detail: string; status: string }
const initialRecords: AdminRecord[] = [
  { id: 'post-reading', section: 'posts', title: 'Make room for a little reading', detail: 'A community note about choosing material that makes you want to come back tomorrow.', status: 'Published' },
  { id: 'post-round5', section: 'posts', title: 'Round 5 reading notes', detail: 'Collect recommendations and reflections from this month’s participants.', status: 'Draft' },
  { id: 'page-about', section: 'pages', title: 'About Tadoku', detail: 'A community built around spending time with the languages you want to understand.', status: 'Published' },
  { id: 'page-guide', section: 'pages', title: 'Start here', detail: 'Choose how you participate, read or listen, then record your activity.', status: 'Published' },
  { id: 'announcement-round5', section: 'announcements', title: 'September Round 5 is open', detail: 'The September round is open. Check the contest schedule for start times and the registration deadline.', status: 'Draft' },
  { id: 'announcement-round6', section: 'announcements', title: 'The November round', detail: 'The next official round begins in November. Check the contest schedule for local dates and times.', status: 'Draft' },
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
const statusTones: Record<string, StatusTagTone> = { Published: 'success', Approved: 'success', Draft: 'warning', Open: 'information', Resolved: 'neutral' }

function RecordEditor({ record, section, onSave, variant }: { record?: AdminRecord; section: AdminSection; variant?: 'default' | 'outline'; onSave: (record: AdminRecord) => void }) {
  const [open, setOpen] = useState(false)
  const review = section === 'reports' || section === 'languages'
  const defaults = { title: record?.title ?? '', detail: record?.detail ?? '', status: record?.status ?? (review ? 'Open' : 'Draft') }
  const methods = usePaperForm({ defaultValues: defaults, onSubmit: ({ value: values }) => { onSave({ ...values, id: record?.id ?? `${section}-${crypto.randomUUID()}`, section }); setOpen(false) } })
  const singular: Record<AdminSection, string> = { posts: 'post', pages: 'page', announcements: 'announcement', languages: 'language', reports: 'report' }
  const title = record ? `${review ? 'Review' : 'Edit'} ${singular[section]}` : `New ${singular[section]}`
  return <Modal title={title} trigger={<Button variant={variant ?? (record ? 'outline' : 'default')} leadingIcon={record ? undefined : <PlusIcon className="paper-icon-default" />}>{record ? review ? 'Review' : 'Edit' : title}</Button>} open={open} onOpenChange={next => { if (next) methods.reset(defaults, { keepDefaultValues: true }); setOpen(next) }} footer={null}>
    <methods.AppForm><methods.Form >
      <div className="app-form__fields">
      <methods.AppField name="title" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => Boolean(String(value).trim()) || 'Enter a title.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.Input label={section === 'languages' ? 'Language or request' : 'Title'} required maxLength={120} />}</methods.AppField>
      <methods.AppField name="detail" validators={{ onChange: ({ value }) => { if (!value) return "This field is required."; const result = (value => Boolean(String(value).trim()) || 'Add the content or review notes.')(value); return result === true ? undefined : result || undefined; } }}>{field => <field.TextArea label={review ? 'Review notes' : 'Content'} required rows={6} />}</methods.AppField>
      <methods.AppField name="status">{field => <field.Select label="Status" options={(review ? ['Open', section === 'languages' ? 'Approved' : 'Resolved'] : ['Draft', 'Published']).map(value => ({ value, label: value }))} />}</methods.AppField>
      </div>
      <div className="app-form__actions"><Button type="submit">Save {review ? 'review' : 'changes'}</Button><Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button></div>
    </methods.Form></methods.AppForm>
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

  if (viewer !== 'admin') return <EmptyState headingLevel={1} title="Admin workspace" description="This workspace is available to administrators." actions={<Button onClick={() => setViewer('admin')}>Enter the demo admin workspace</Button>} />

  return <div className="records-admin-layout">
    <aside className="records-admin-sidebar">{sidebar}<Link className="records-admin-return text-link" to="/">Back to Tadoku</Link></aside>
    <div className="records-admin-main">
      <div className="records-admin-mobile"><Drawer trigger={<Button variant="outline" leadingIcon={<Bars3Icon className="paper-icon-default" />}>Admin sections</Button>} title="Admin workspace" placement="start" open={mobileOpen} onOpenChange={setMobileOpen}>{sidebar}<Link className="text-link" to="/" onClick={() => setMobileOpen(false)}>Back to Tadoku</Link></Drawer></div>
      <Breadcrumb items={[{ id: 'home', label: 'Tadoku', href: '/' }, { id: 'admin', label: 'Admin', href: '/admin' }, { id: section, label: names[section] ?? 'Page not found' }]} renderLink={renderLink} />
      <PageHeader className="records-admin-heading" title={names[section] ?? 'Page not found'} actions={section === 'dashboard' ? <RecordEditor section="announcements" onSave={saveRecord} /> : ['posts', 'pages', 'announcements', 'languages'].includes(section) ? <RecordEditor section={section as AdminSection} onSave={saveRecord} /> : null} />
      {notice ? <Flash variant="success" className="records-admin-notice">{notice}</Flash> : null}
      {section === 'dashboard' ? <>
        <StatGrid className="records-admin-stats" stats={[
          { label: 'Community members', value: users.length },
          { label: 'Active official round', value: 'Round 5', detail: currentContest ? formatDateRange(contestStart(currentContest.start), contestEnd(currentContest.end)) : 'Schedule unavailable' },
          { label: 'Open reports', value: openReports.length },
        ]} />
        <section className="page-section records-admin-attention"><SectionHeading level={2} title="Needs attention" actions={<Link className="text-link" to="/admin/reports">View reports</Link>} /><ul className="records-admin-queue">
          {openReports.map(record => <li key={record.id}><div><strong>{record.title}</strong><p>{record.detail}</p></div><RecordEditor record={record} section="reports" onSave={saveRecord} /></li>)}
          {records.filter(record => record.section === 'announcements' && record.status === 'Draft').slice(0, 1).map(record => <li key={record.id}><div><strong>{record.title}</strong><p>Draft announcement ready for an editorial check.</p></div><RecordEditor record={record} section="announcements" onSave={saveRecord} /></li>)}
          {requestedLanguages.length ? <li><div><strong>{requestedLanguages.length} language requests</strong><p>Check names and supported-language entries before approval.</p></div><Link className={buttonClassName({ variant: 'outline' })} to="/admin/languages">Review requests</Link></li> : null}
          {!openReports.length && !requestedLanguages.length ? <li><p>All reports and language requests have been reviewed.</p></li> : null}
        </ul></section>
        <div className="records-admin-bottom">
          <section className="page-section"><SectionHeading level={2} title="Current contest" description={currentContest?.title ?? '2026 Round 5'} actions={<Link className="text-link" to="/contests/round5">Open contest</Link>} /><DescriptionList layout="inline" items={[
            { label: 'Period', value: currentContest ? formatDateRange(contestStart(currentContest.start), contestEnd(currentContest.end)) : 'Schedule unavailable' },
            { label: 'Registration deadline', value: currentContest ? formatDateTime(contestEnd(currentContest.registrationDeadline)) : 'Schedule unavailable' },
            { label: 'Submitted sample entries', value: logs.filter(log => log.submissions.some(item => item.contestId === 'round5')).length },
          ]} /></section>
          <section className="page-section"><SectionHeading level={2} title="Quick actions" /><div className="records-admin-quick"><RecordEditor section="posts" variant="outline" onSave={saveRecord} /><RecordEditor section="pages" variant="outline" onSave={saveRecord} /><Link className={buttonClassName({ variant: 'outline' })} to="/admin/languages">Languages</Link><Link className={buttonClassName({ variant: 'outline' })} to="/admin/users">Users</Link></div></section>
        </div>
        <section className="page-section records-admin-status"><SectionHeading level={2} title="Service status" description="Illustrative service states for this unpublished preview." /><DescriptionList layout="inline" items={['Immersion API', 'Content API', 'Background jobs'].map(service => ({ label: service, value: <StatusTag tone="success">Healthy sample</StatusTag> }))} /></section>
      </> : section === 'users' ? <Table className="records-admin-table" caption="Community members" captionVisibility="screen-reader" minWidth="32rem" columns={[{ id: 'name', header: 'Member', rowHeader: true, cell: item => <Link className="text-link" to={`/users/${item.id}`}>{item.name}</Link> }, { id: 'joined', header: 'Tracking since', cell: item => formatDate(item.joined) }, { id: 'action', header: 'Activity', cell: item => <Link className="text-link" to={`/users/${item.id}/activity`}>View activity</Link> }]} rows={users} getRowKey={item => item.id} /> : section === 'contests' ? <Table className="records-admin-table" caption="Contests" captionVisibility="screen-reader" minWidth="38rem" columns={[{ id: 'title', header: 'Contest', rowHeader: true, cell: item => <Link className="text-link" to={`/contests/${item.id}`}>{item.title}</Link> }, { id: 'type', header: 'Collection', cell: item => item.unlisted ? 'Unlisted' : item.scope === 'official' ? 'Official' : 'Community' }, { id: 'dates', header: 'Begins', cell: item => formatDate(contestStart(item.start)) }, { id: 'action', header: 'Details', cell: item => <Link className="text-link" to={`/contests/${item.id}`}>Open contest</Link> }]} rows={contests} getRowKey={item => item.id} /> : ['posts', 'pages', 'announcements', 'languages', 'reports'].includes(section) ? <section className="records-admin-collection" aria-label={names[section]}>
        <p className="records-admin-count">{records.filter(record => record.section === section).length} {names[section].toLocaleLowerCase()}</p>
        <ul className="records-admin-records">{records.filter(record => record.section === section).map(item => <li key={item.id}>
          <div className="records-admin-record"><h2>{item.title}</h2><p>{item.detail}</p></div>
          <StatusTag tone={statusTones[item.status] ?? 'neutral'}>{item.status}</StatusTag>
          <RecordEditor record={item} section={item.section} onSave={saveRecord} />
        </li>)}</ul>
        {!records.some(record => record.section === section) ? <EmptyState headingLevel={2} title={`No ${names[section].toLocaleLowerCase()} yet.`} /> : null}
      </section> : <EmptyState headingLevel={2} title="This admin page isn’t available" actions={<Link className={buttonClassName()} to="/admin">Return to dashboard</Link>} />}
    </div>
  </div>
}
