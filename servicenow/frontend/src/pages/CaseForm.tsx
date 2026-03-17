import { useEffect, useState, useCallback, useRef, useImperativeHandle, forwardRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api, type CaseDetail } from '../api/client'

const USER_ID = 'frontend-user'

const WORKFLOW_STAGES = ['Draft', 'Analysis', 'Contain', 'Eradicate', 'Recover', 'Review', 'Resolved', 'Closed']

export function CaseForm() {
  const { caseId } = useParams<{ caseId: string }>()
  const navigate = useNavigate()
  const incidentFormRef = useRef<{ save: (redirectAfter: boolean) => void } | null>(null)
  const [detail, setDetail] = useState<CaseDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [relatedTab, setRelatedTab] = useState<'observables' | 'similarIncidents' | 'worknotes' | 'audit' | 'attachments' | 'alerts' | 'enrichment'>('observables')

  const load = useCallback(() => {
    if (!caseId) return
    setLoading(true)
    setError(null)
    api
      .getCaseDetail(caseId, USER_ID)
      .then(setDetail)
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [caseId])

  useEffect(() => load(), [load])

  if (loading) {
    return (
      <ServicenowLayout contextTitle="Security Incident">
        <main className="app-main"><div className="sn-page">Loading…</div></main>
      </ServicenowLayout>
    )
  }
  if (error) {
    return (
      <ServicenowLayout contextTitle="Security Incident">
        <main className="app-main"><div className="sn-page" style={{ color: '#cf222e' }}>{error}</div></main>
      </ServicenowLayout>
    )
  }
  if (!detail?.case) return null

  const c = detail.case
  const caseNumber = c.case_number || c.id
  const currentStage = WORKFLOW_STAGES.includes(c.state) ? c.state : 'Draft'

  return (
    <ServicenowLayout contextTitle="Security Incident">
      <div className="sn-case-bar">
        <button type="button" className="sn-case-bar-back" onClick={() => navigate('/cases')} aria-label="Back">
          ‹
        </button>
        <span className="sn-case-bar-title">Security Incident: {caseNumber}</span>
        <div className="sn-case-bar-actions">
          <button
            type="button"
            className="sn-btn sn-btn-primary"
            onClick={() => incidentFormRef.current?.save(false)}
          >
            Save
          </button>
          <button
            type="button"
            className="sn-btn sn-btn-primary"
            onClick={() => incidentFormRef.current?.save(true)}
          >
            Update
          </button>
        </div>
      </div>
      <div className="sn-incident-layout">
        <div className="sn-main" style={{ width: '100%' }}>
          <div className="sn-content-proportional">
          <div style={{ marginBottom: '0.5rem' }}>
            <Link to="/cases" style={{ fontSize: 13 }}>← Back to list</Link>
          </div>
          <h1 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '0.5rem' }}>Security Incident {caseNumber}</h1>

          <div className="sn-workflow-stages">
            {WORKFLOW_STAGES.map((s) => (
              <span key={s} className={`sn-stage ${s === currentStage ? 'active' : ''}`}>{s}</span>
            ))}
          </div>

          <IncidentDetailsForm
            ref={incidentFormRef}
            caseId={caseId!}
            detail={detail}
            onUpdated={load}
            onSavedAndRedirect={() => navigate('/cases')}
          />

          <div className="sn-tabs" style={{ marginTop: '1.5rem' }}>
            <button type="button" className={`sn-tab ${relatedTab === 'observables' ? 'active' : ''}`} onClick={() => setRelatedTab('observables')}>Associated Observables ({detail.observables?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'similarIncidents' ? 'active' : ''}`} onClick={() => setRelatedTab('similarIncidents')}>Similar Security Incidents ({detail.similar_incidents?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'worknotes' ? 'active' : ''}`} onClick={() => setRelatedTab('worknotes')}>Work Notes ({detail.worknotes?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'attachments' ? 'active' : ''}`} onClick={() => setRelatedTab('attachments')}>Attachments ({detail.attachments?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'alerts' ? 'active' : ''}`} onClick={() => setRelatedTab('alerts')}>Alerts ({detail.alerts?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'enrichment' ? 'active' : ''}`} onClick={() => setRelatedTab('enrichment')}>Enrichment Results ({detail.enrichment_results?.length ?? 0})</button>
            <button type="button" className={`sn-tab ${relatedTab === 'audit' ? 'active' : ''}`} onClick={() => setRelatedTab('audit')}>Audit History ({detail.audit_events?.length ?? 0})</button>
          </div>
          {relatedTab === 'observables' && <RelatedListObservables caseId={caseId!} detail={detail} onUpdated={load} />}
          {relatedTab === 'similarIncidents' && <RelatedListSimilarIncidents detail={detail} />}
          {relatedTab === 'worknotes' && <RelatedListWorknotes caseId={caseId!} detail={detail} onUpdated={load} />}
          {relatedTab === 'attachments' && <RelatedListAttachments caseId={caseId!} detail={detail} onUpdated={load} />}
          {relatedTab === 'alerts' && <RelatedListAlerts detail={detail} />}
          {relatedTab === 'enrichment' && <RelatedListEnrichment caseId={caseId!} detail={detail} />}
          {relatedTab === 'audit' && <RelatedListAudit detail={detail} />}
          </div>
        </div>
      </div>
    </ServicenowLayout>
  )
}

const PRIORITY_OPTIONS = ['-- None --', '1 - Critical', '2 - High', '3 - Medium', '4 - Low']
const SEVERITY_OPTIONS = ['-- None --', '1 - Critical', '2 - High', '3 - Medium', '4 - Low']
const ENV_LEVEL_OPTIONS = ['-- None --', 'Dev', 'Test', 'Model', 'Prod']
const ENV_TYPE_OPTIONS = ['-- None --', 'Internal', 'Diverstitute', 'External']
const IMPACTED_OBJECT_OPTIONS = ['-- None --', 'Network', 'Compute', 'Policy/Config', 'Identity', 'Data', 'Application/Service']
const SOURCE_OPTIONS = ['-- None --', 'Alert', 'Request', 'Notification', 'CTR Finding']
const CATEGORY_OPTIONS = ['-- None --', ...Array.from({ length: 13 }, (_, i) => `CAT ${i}`)]
const CSIRT_OPTIONS = ['-- None --', 'CSIRT Consultation', 'CSIRT EOI', 'CSIRT Incident']
const severityToDisplay = (s: string) => ({ critical: '1 - Critical', high: '2 - High', medium: '3 - Medium', low: '4 - Low' }[(s || '').toLowerCase()] ?? '-- None --')
const priorityToDisplay = (p: string) => (p || '').match(/^P?1$/i) ? '1 - Critical' : (p || '').match(/^P?2$/i) ? '2 - High' : (p || '').match(/^P?3$/i) ? '3 - Medium' : (p || '').match(/^P?4$/i) ? '4 - Low' : '-- None --'

const IncidentDetailsForm = forwardRef<
  { save: (redirectAfter: boolean) => void },
  { caseId: string; detail: CaseDetail; onUpdated: () => void; onSavedAndRedirect?: () => void }
>(function IncidentDetailsForm({ caseId, detail, onUpdated, onSavedAndRedirect }, ref) {
  const c = detail.case
  const formId = `incident-details-form-${caseId}`
  const [detailsTab, setDetailsTab] = useState<'incident' | 'closure'>('incident')
  const redirectAfterSaveRef = useRef(false)
  const [title, setTitle] = useState(c.title)
  const [state, setState] = useState(c.state || 'Draft')
  const [priority, setPriority] = useState(c.priority)
  const [severity, setSeverity] = useState(severityToDisplay(c.severity ?? ''))
  const [assignedUserId, setAssignedUserId] = useState(c.assigned_user_id ?? '')
  const [assignmentGroupId, setAssignmentGroupId] = useState(c.assignment_group_id ?? '')
  const [affectedUserId, setAffectedUserId] = useState(c.affected_user_id ?? '')
  const [followupTime, setFollowupTime] = useState(c.followup_time ? c.followup_time.slice(0, 16) : '')
  const [category, setCategory] = useState(c.category ?? '')
  const [subcategory, setSubcategory] = useState(c.subcategory ?? '')
  const [source, setSource] = useState(c.source ?? '')
  const [sourceTool, setSourceTool] = useState(c.source_tool ?? '')
  const [sourceToolFeature, setSourceToolFeature] = useState(c.source_tool_feature ?? '')
  const [configurationItem, setConfigurationItem] = useState(c.configuration_item ?? '')
  const [socNotes, setSocNotes] = useState(c.soc_notes ?? '')
  const [nextSteps, setNextSteps] = useState(c.next_steps ?? '')
  const [csirtClassification, setCsirtClassification] = useState(c.csirt_classification ?? '')
  const [socLeadUserId, setSocLeadUserId] = useState(c.soc_lead_user_id ?? '')
  const [requestedByUserId, setRequestedByUserId] = useState(c.requested_by_user_id ?? '')
  const [environmentLevel, setEnvironmentLevel] = useState(c.environment_level ?? '')
  const [environmentType, setEnvironmentType] = useState(c.environment_type ?? '')
  const [pdn, setPdn] = useState(c.pdn ?? '')
  const [impactedObject, setImpactedObject] = useState(c.impacted_object ?? '')
  const [notificationTime, setNotificationTime] = useState(c.notification_time ? c.notification_time.slice(0, 16) : '')
  const [eventOccurredTime, setEventOccurredTime] = useState(c.event_occurred_time ? c.event_occurred_time.slice(0, 16) : '')
  const [mttr, setMttr] = useState(c.mttr ?? '')
  const [reassignmentCount, setReassignmentCount] = useState(String(c.reassignment_count ?? 0))
  const [assignedToCount, setAssignedToCount] = useState(String(c.assigned_to_count ?? 0))
  const [isAffectedUserVip, setIsAffectedUserVip] = useState(c.is_affected_user_vip ?? false)
  const [engineeringDocument, setEngineeringDocument] = useState(c.engineering_document ?? '')
  const [responseDocument, setResponseDocument] = useState(c.response_document ?? '')
  const [sourceAccuracy, setSourceAccuracy] = useState(c.accuracy ?? '-- None --')
  const [determination, setDetermination] = useState(c.determination ?? '-- None --')
  const [closureReason, setClosureReason] = useState(c.closure_reason ?? '')
  const [timeWorked, setTimeWorked] = useState('')
  const [users, setUsers] = useState<{ id: string; display_name?: string }[]>([])
  const [groups, setGroups] = useState<{ id: string; name?: string }[]>([])
  const [usersError, setUsersError] = useState<string | null>(null)
  const [groupsError, setGroupsError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [updateError, setUpdateError] = useState<string | null>(null)

  useEffect(() => {
    setUsersError(null)
    setGroupsError(null)
    api.listUsers(USER_ID).then((r) => setUsers(r.users || [])).catch((err) => setUsersError(err instanceof Error ? err.message : 'Failed to load users'))
    api.listGroups(USER_ID).then((r) => setGroups(r.groups || [])).catch((err) => setGroupsError(err instanceof Error ? err.message : 'Failed to load assignment groups'))
  }, [])
  useEffect(() => {
    setTitle(c.title)
    setState(c.state || 'Draft')
    setPriority(priorityToDisplay(c.priority ?? ''))
    setUpdateError(null)
    setSeverity(severityToDisplay(c.severity ?? ''))
    setAssignedUserId(c.assigned_user_id ?? '')
    setAssignmentGroupId(c.assignment_group_id ?? '')
    setAffectedUserId(c.affected_user_id ?? '')
    setFollowupTime(c.followup_time ? c.followup_time.slice(0, 16) : '')
    setNotificationTime(c.notification_time ? c.notification_time.slice(0, 16) : '')
    setEventOccurredTime(c.event_occurred_time ? c.event_occurred_time.slice(0, 16) : '')
    setCategory(c.category ?? '')
    setSubcategory(c.subcategory ?? '')
    setSource(c.source ?? '')
    setSourceTool(c.source_tool ?? '')
    setSourceToolFeature(c.source_tool_feature ?? '')
    setConfigurationItem(c.configuration_item ?? '')
    setSocNotes(c.soc_notes ?? '')
    setNextSteps(c.next_steps ?? '')
    setCsirtClassification(c.csirt_classification ?? '')
    setSocLeadUserId(c.soc_lead_user_id ?? '')
    setRequestedByUserId(c.requested_by_user_id ?? '')
    setEnvironmentLevel(c.environment_level ?? '')
    setEnvironmentType(c.environment_type ?? '')
    setPdn(c.pdn ?? '')
    setImpactedObject(c.impacted_object ?? '')
    setMttr(c.mttr ?? '')
    setReassignmentCount(String(c.reassignment_count ?? 0))
    setAssignedToCount(String(c.assigned_to_count ?? 0))
    setIsAffectedUserVip(c.is_affected_user_vip ?? false)
    setEngineeringDocument(c.engineering_document ?? '')
    setResponseDocument(c.response_document ?? '')
    setSourceAccuracy(c.accuracy ?? '-- None --')
    setDetermination(c.determination ?? '-- None --')
    setClosureReason(c.closure_reason ?? '')
  }, [c.id, c.title, c.state, c.priority, c.severity, c.assigned_user_id, c.assignment_group_id, c.affected_user_id, c.followup_time, c.notification_time, c.event_occurred_time, c.category, c.subcategory, c.source, c.source_tool, c.source_tool_feature, c.configuration_item, c.soc_notes, c.next_steps, c.csirt_classification, c.soc_lead_user_id, c.requested_by_user_id, c.environment_level, c.environment_type, c.pdn, c.impacted_object, c.mttr, c.reassignment_count, c.assigned_to_count, c.is_affected_user_vip, c.engineering_document, c.response_document, c.accuracy, c.determination, c.closure_reason])
  useEffect(() => {
    const opened = c.opened_time ? new Date(c.opened_time).getTime() : null
    if (!opened) {
      setTimeWorked('')
      return
    }
    const tick = () => {
      const sec = Math.floor((Date.now() - opened) / 1000)
      const h = Math.floor(sec / 3600)
      const m = Math.floor((sec % 3600) / 60)
      const s = sec % 60
      setTimeWorked(`${h}h ${m}m ${s}s`)
    }
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [c.opened_time])

  const priorityToApi = (p: string) => (p === '1 - Critical' ? 'P1' : p === '2 - High' ? 'P2' : p === '3 - Medium' ? 'P3' : p === '4 - Low' ? 'P4' : undefined)
  const severityToApi = (s: string) => (s === '1 - Critical' ? 'critical' : s === '2 - High' ? 'high' : s === '3 - Medium' ? 'medium' : s === '4 - Low' ? 'low' : undefined)
  const performSave = useCallback(async () => {
    setSaving(true)
    setUpdateError(null)
    try {
      await api.updateCase(caseId, {
        title,
        state,
        priority: priorityToApi(priority),
        severity: severityToApi(severity),
        assigned_user_id: assignedUserId || undefined,
        assignment_group_id: assignmentGroupId || undefined,
        affected_user_id: affectedUserId || undefined,
        followup_time: followupTime ? new Date(followupTime).toISOString() : undefined,
        notification_time: notificationTime ? new Date(notificationTime).toISOString() : undefined,
        event_occurred_time: eventOccurredTime ? new Date(eventOccurredTime).toISOString() : undefined,
        category: category || undefined,
        subcategory: subcategory || undefined,
        source: source || undefined,
        source_tool: sourceTool || undefined,
        source_tool_feature: sourceToolFeature || undefined,
        configuration_item: configurationItem || undefined,
        soc_notes: socNotes || undefined,
        next_steps: nextSteps || undefined,
        csirt_classification: csirtClassification || undefined,
        soc_lead_user_id: socLeadUserId || undefined,
        requested_by_user_id: requestedByUserId || undefined,
        environment_level: environmentLevel || undefined,
        environment_type: environmentType || undefined,
        pdn: pdn || undefined,
        impacted_object: impactedObject || undefined,
        mttr: mttr || undefined,
        reassignment_count: parseInt(reassignmentCount, 10) || 0,
        assigned_to_count: parseInt(assignedToCount, 10) || 0,
        is_affected_user_vip: isAffectedUserVip,
        engineering_document: engineeringDocument || undefined,
        response_document: responseDocument || undefined,
        accuracy: sourceAccuracy === '-- None --' ? undefined : sourceAccuracy,
        determination: determination === '-- None --' ? undefined : determination,
        closure_reason: closureReason?.trim() || undefined,
      }, USER_ID)
      setUpdateError(null)
      onUpdated()
      if (redirectAfterSaveRef.current && onSavedAndRedirect) {
        onSavedAndRedirect()
      }
    } catch (err) {
      setUpdateError(err instanceof Error ? err.message : 'Update failed')
    } finally {
      setSaving(false)
    }
  }, [caseId, title, state, priority, severity, assignedUserId, assignmentGroupId, affectedUserId, followupTime, notificationTime, eventOccurredTime, category, subcategory, source, sourceTool, sourceToolFeature, configurationItem, socNotes, nextSteps, csirtClassification, socLeadUserId, requestedByUserId, environmentLevel, environmentType, pdn, impactedObject, mttr, reassignmentCount, assignedToCount, isAffectedUserVip, engineeringDocument, responseDocument, sourceAccuracy, determination, closureReason, onUpdated, onSavedAndRedirect])

  useImperativeHandle(ref, () => ({
    save(redirectAfter: boolean) {
      redirectAfterSaveRef.current = redirectAfter
      void performSave()
    },
  }), [performSave])

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    redirectAfterSaveRef.current = false
    await performSave()
  }

  const readOnlyFieldStyle: React.CSSProperties = {
    width: '100%',
    padding: '0.35rem 0.5rem',
    backgroundColor: 'var(--sn-bg-disabled, #f5f5f5)',
    color: 'var(--sn-text-muted, #666)',
    border: '1px solid var(--sn-border, #ddd)',
    cursor: 'not-allowed',
  }

  return (
    <div style={{ background: '#fff', border: '1px solid var(--sn-border)', padding: '1rem', marginBottom: '1rem' }}>
      <form id={formId} onSubmit={handleUpdate}>
        {updateError && (
          <div className="sn-form-error" style={{ marginBottom: '1rem' }}>{updateError}</div>
        )}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sn-form-group">
              <label>State</label>
              <select value={state} onChange={(e) => setState(e.target.value)} disabled={saving || (c.state === 'Closed' || c.state === 'Resolved')} style={{ padding: '0.35rem 0.5rem', minWidth: 140 }}>
                {WORKFLOW_STAGES.map((s) => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Followup date and time</label>
              <input type="datetime-local" value={followupTime} onChange={(e) => setFollowupTime(e.target.value)} disabled={saving} placeholder=" " style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Requested By</label>
              <select value={requestedByUserId} onChange={(e) => setRequestedByUserId(e.target.value)} disabled={saving || !!usersError} style={{ width: '100%' }}>
                <option value="">— None —</option>
                {users.map((u) => <option key={u.id} value={u.id}>{u.display_name || u.id}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Environment Level</label>
              <select value={environmentLevel} onChange={(e) => setEnvironmentLevel(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {ENV_LEVEL_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Environment Type</label>
              <select value={environmentType} onChange={(e) => setEnvironmentType(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {ENV_TYPE_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>PDN</label>
              <input type="text" value={pdn} onChange={(e) => setPdn(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Impacted Object</label>
              <select value={impactedObject} onChange={(e) => setImpactedObject(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {IMPACTED_OBJECT_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Source</label>
              <select value={source} onChange={(e) => setSource(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {SOURCE_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Configuration item</label>
              <input type="text" value={configurationItem} onChange={(e) => setConfigurationItem(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Affected User</label>
              <select value={affectedUserId} onChange={(e) => setAffectedUserId(e.target.value)} disabled={saving || !!usersError} style={{ width: '100%' }}>
                <option value="">— None —</option>
                {users.map((u) => <option key={u.id} value={u.id}>{u.display_name || u.id}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Category</label>
              <select value={category} onChange={(e) => setCategory(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {CATEGORY_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Sub Category</label>
              <input type="text" value={subcategory} onChange={(e) => setSubcategory(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sn-form-group">
              <label>Number</label>
              <input type="text" value={c.case_number || '—'} readOnly disabled style={readOnlyFieldStyle} />
            </div>
            <div className="sn-form-group">
              <label>Opened By</label>
              <input type="text" value={users.find((u) => u.id === c.opened_by_user_id)?.display_name || c.opened_by_user_id || '—'} readOnly disabled style={readOnlyFieldStyle} />
            </div>
            <div className="sn-form-group">
              <label>Opened</label>
              <input type="text" value={c.opened_time ? new Date(c.opened_time).toLocaleString() : '—'} readOnly disabled style={readOnlyFieldStyle} />
            </div>
            <div className="sn-form-group">
              <label>Notification Time</label>
              <input type="datetime-local" value={notificationTime} onChange={(e) => setNotificationTime(e.target.value)} disabled={saving} placeholder=" " style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Event time</label>
              <input type="datetime-local" value={eventOccurredTime} onChange={(e) => setEventOccurredTime(e.target.value)} disabled={saving} placeholder=" " style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>MTTR</label>
              <input type="text" value={mttr} onChange={(e) => setMttr(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Priority</label>
              <select value={priority} onChange={(e) => setPriority(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {PRIORITY_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Severity</label>
              <select value={severity} onChange={(e) => setSeverity(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {SEVERITY_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Assignment group</label>
              {groupsError && <div className="sn-form-error" style={{ marginBottom: 4 }}>{groupsError}</div>}
              <select value={assignmentGroupId} onChange={(e) => setAssignmentGroupId(e.target.value)} disabled={saving || !!groupsError} style={{ width: '100%' }}>
                <option value="">— None —</option>
                {groups.map((g) => <option key={g.id} value={g.id}>{g.name || g.id}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Assigned to</label>
              <select value={assignedUserId} onChange={(e) => setAssignedUserId(e.target.value)} disabled={saving || !!usersError} style={{ width: '100%' }}>
                <option value="">— None —</option>
                {users.map((u) => <option key={u.id} value={u.id}>{u.display_name || u.id}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Reassignment count</label>
              <input type="number" min={0} value={reassignmentCount} onChange={(e) => setReassignmentCount(e.target.value)} disabled={saving} style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>Assigned to count</label>
              <input type="number" min={0} value={assignedToCount} onChange={(e) => setAssignedToCount(e.target.value)} disabled={saving} style={{ width: '100%' }} />
            </div>
            <div className="sn-form-group">
              <label>SOC Lead</label>
              <select value={socLeadUserId} onChange={(e) => setSocLeadUserId(e.target.value)} disabled={saving || !!usersError} style={{ width: '100%' }}>
                <option value="">— None —</option>
                {users.map((u) => <option key={u.id} value={u.id}>{u.display_name || u.id}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>CSIRT Classification</label>
              <select value={csirtClassification} onChange={(e) => setCsirtClassification(e.target.value)} disabled={saving} style={{ width: '100%' }}>
                {CSIRT_OPTIONS.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div className="sn-form-group">
              <label>Is Affected User A VIP</label>
              <select value={isAffectedUserVip ? 'Yes' : 'No'} onChange={(e) => setIsAffectedUserVip(e.target.value === 'Yes')} disabled={saving} style={{ width: '100%' }}>
                <option value="No">No</option>
                <option value="Yes">Yes</option>
              </select>
            </div>
            <div className="sn-form-group">
              <label>Time worked</label>
              <input type="text" value={timeWorked || '—'} readOnly disabled style={{ ...readOnlyFieldStyle, fontVariantNumeric: 'tabular-nums' }} />
            </div>
          </div>
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <div className="sn-form-group" style={{ marginBottom: '0.75rem' }}>
            <label>Short Description <span className="required">*</span></label>
            <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} disabled={saving} style={{ width: '100%' }} />
          </div>
          <div className="sn-form-group" style={{ marginBottom: '0.75rem' }}>
            <label>Engineering document</label>
            <input type="text" value={engineeringDocument} onChange={(e) => setEngineeringDocument(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
          </div>
          <div className="sn-form-group">
            <label>Response Document</label>
            <input type="text" value={responseDocument} onChange={(e) => setResponseDocument(e.target.value)} disabled={saving} placeholder="—" style={{ width: '100%' }} />
          </div>
        </div>
      </form>
      <div className="sn-tabs" style={{ marginTop: '1.5rem' }}>
        <button type="button" className={`sn-tab ${detailsTab === 'incident' ? 'active' : ''}`} onClick={() => setDetailsTab('incident')}>Incident details</button>
        <button type="button" className={`sn-tab ${detailsTab === 'closure' ? 'active' : ''}`} onClick={() => setDetailsTab('closure')}>Closure Information</button>
      </div>
      {detailsTab === 'incident' && <WorkNotesBlock caseId={caseId} detail={detail} users={users} onUpdated={onUpdated} />}
      {detailsTab === 'closure' && (
        <ClosureInformationBlock
          detail={detail}
          users={users}
          sourceAccuracy={sourceAccuracy}
          setSourceAccuracy={setSourceAccuracy}
          determination={determination}
          setDetermination={setDetermination}
          closureReason={closureReason}
          setClosureReason={setClosureReason}
          saving={saving}
        />
      )}
      <button type="submit" form={formId} className="sn-btn sn-btn-primary" disabled={saving} style={{ marginTop: '1rem' }}>{saving ? 'Saving…' : 'Update'}</button>
    </div>
  )
})

const SOURCE_ACCURACY_OPTIONS = ['-- None --', 'True Positive', 'False Positive']
const DETERMINATION_OPTIONS = ['-- None --', 'Unknown', 'Malicious', 'Suspicious', 'Clean']

/** Only in Review state can closure information be entered and saved. */
const CLOSURE_ALLOWED_STATES = ['Review']
/** In Closed (and Resolved) state, closure information is shown read-only. */
const CLOSURE_READONLY_STATES = ['Closed', 'Resolved']

function ClosureInformationBlock({
  detail,
  users,
  sourceAccuracy,
  setSourceAccuracy,
  determination,
  setDetermination,
  closureReason,
  setClosureReason,
  saving,
}: {
  detail: CaseDetail
  users: { id: string; display_name?: string }[]
  sourceAccuracy: string
  setSourceAccuracy: (v: string) => void
  determination: string
  setDetermination: (v: string) => void
  closureReason: string
  setClosureReason: (v: string) => void
  saving: boolean
}) {
  const c = detail.case
  const state = (c.state || '').trim()
  const canEditClosure = CLOSURE_ALLOWED_STATES.some((s) => state === s)
  const isReadOnlyClosure = CLOSURE_READONLY_STATES.some((s) => state === s)

  const resolvedByDisplay = c.closed_by_user_id
    ? (users.find((u) => u.id === c.closed_by_user_id)?.display_name || c.closed_by_user_id)
    : '—'
  const resolvedDateDisplay = c.closed_time
    ? new Date(c.closed_time).toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' })
    : '—'

  // Not in Review and not Closed/Resolved: show message to move to Review
  if (!canEditClosure && !isReadOnlyClosure) {
    return (
      <div style={{ background: '#fff', border: '1px solid var(--sn-border)', padding: '1rem', marginBottom: '1rem' }}>
        <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>Closure Information</h3>
        <div style={{ color: '#666', fontSize: 14 }}>
          <p>Closure information can only be entered when the security incident is in <strong>Review</strong> state.</p>
          <p>Current state: <strong>{state || '—'}</strong>. Update the incident state to Review on the Incident details tab, then return here to enter closure information. Use the Update button below to save.</p>
        </div>
      </div>
    )
  }

  // Greyed-out style for read-only fields (look like inputs but not editable)
  const readOnlyInputStyle: React.CSSProperties = {
    width: '100%',
    padding: '0.35rem 0.5rem',
    backgroundColor: 'var(--sn-bg-disabled, #f5f5f5)',
    color: 'var(--sn-text-muted, #666)',
    border: '1px solid var(--sn-border, #ddd)',
    cursor: 'not-allowed',
  }
  const readOnlyTextareaStyle: React.CSSProperties = {
    ...readOnlyInputStyle,
    minHeight: '5rem',
    resize: 'vertical',
  }

  // Closed or Resolved: show closure information read-only (fixed to what was set, greyed-out inputs)
  if (isReadOnlyClosure) {
    return (
      <div style={{ background: '#fff', border: '1px solid var(--sn-border)', padding: '1rem', marginBottom: '1rem' }}>
        <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>Closure Information</h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sn-form-group">
              <label>Source Accuracy</label>
              <input type="text" value={c.accuracy || '—'} readOnly disabled style={readOnlyInputStyle} />
            </div>
            <div className="sn-form-group">
              <label>Determination</label>
              <input type="text" value={c.determination || '—'} readOnly disabled style={readOnlyInputStyle} />
            </div>
            <div className="sn-form-group">
              <label>Resolved By</label>
              <input type="text" value={resolvedByDisplay} readOnly disabled style={readOnlyInputStyle} />
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sn-form-group">
              <label>Resolved Date</label>
              <input type="text" value={resolvedDateDisplay} readOnly disabled style={readOnlyInputStyle} />
            </div>
          </div>
        </div>
        <div className="sn-form-group" style={{ marginBottom: 0 }}>
          <label>Close Notes</label>
          <textarea value={c.closure_reason || '—'} readOnly disabled rows={4} style={readOnlyTextareaStyle} />
        </div>
      </div>
    )
  }

  // Review: editable closure form
  return (
    <div style={{ background: '#fff', border: '1px solid var(--sn-border)', padding: '1rem', marginBottom: '1rem' }}>
      <h3 style={{ fontSize: '1rem', marginBottom: '1rem' }}>Closure Information</h3>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <div className="sn-form-group">
            <label>Source Accuracy <span className="required">*</span></label>
            <select value={sourceAccuracy} onChange={(e) => setSourceAccuracy(e.target.value)} disabled={saving} style={{ padding: '0.35rem 0.5rem', minWidth: 160 }}>
              {SOURCE_ACCURACY_OPTIONS.map((opt) => <option key={opt} value={opt}>{opt}</option>)}
            </select>
          </div>
          <div className="sn-form-group">
            <label>Determination <span className="required">*</span></label>
            <select value={determination} onChange={(e) => setDetermination(e.target.value)} disabled={saving} style={{ padding: '0.35rem 0.5rem', minWidth: 160 }}>
              {DETERMINATION_OPTIONS.map((opt) => <option key={opt} value={opt}>{opt}</option>)}
            </select>
          </div>
          <div className="sn-form-group">
            <label>Resolved By</label>
            <input type="text" value={resolvedByDisplay} readOnly disabled style={{ width: '100%', padding: '0.35rem 0.5rem', backgroundColor: 'var(--sn-bg-disabled, #f5f5f5)', color: 'var(--sn-text-muted, #666)', border: '1px solid var(--sn-border, #ddd)', cursor: 'not-allowed' }} />
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <div className="sn-form-group">
            <label>Resolved Date</label>
            <input type="text" value={resolvedDateDisplay} readOnly disabled style={{ width: '100%', padding: '0.35rem 0.5rem', backgroundColor: 'var(--sn-bg-disabled, #f5f5f5)', color: 'var(--sn-text-muted, #666)', border: '1px solid var(--sn-border, #ddd)', cursor: 'not-allowed' }} />
          </div>
        </div>
      </div>
      <div className="sn-form-group" style={{ marginBottom: 0 }}>
        <label>Close Notes <span className="required">*</span></label>
        <textarea value={closureReason} onChange={(e) => setClosureReason(e.target.value)} disabled={saving} rows={4} placeholder="Required when saving closure. Use Update button below to save." style={{ width: '100%' }} />
      </div>
    </div>
  )
}

function WorkNotesBlock({ caseId, detail, users, onUpdated }: { caseId: string; detail: CaseDetail; users: { id: string; display_name?: string }[]; onUpdated: () => void }) {
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const worknotes = detail.worknotes ?? []

  const handlePost = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    setSubmitting(true)
    setError(null)
    try {
      await api.addWorknote(caseId, content.trim(), USER_ID)
      setContent('')
      onUpdated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to post work note')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="sn-work-notes-section">
      <div className="sn-wn-tabs">
        <span className="sn-wn-tab active">Work Notes</span>
      </div>
      <div className="sn-wn-body">
        <form onSubmit={handlePost}>
          {error && <div className="sn-form-error" style={{ marginBottom: '0.5rem' }}>{error}</div>}
          <textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder="Add work note…" disabled={submitting} />
          <button type="submit" className="sn-btn-post" disabled={submitting || !content.trim()}>Post</button>
        </form>
      </div>
      <div className="sn-activities">
        <div className="sn-activities-header">Activities: {worknotes.length}</div>
        {worknotes.length === 0 ? (
          <div className="sn-activity-item sn-act-meta">No activities yet.</div>
        ) : (
          worknotes.slice(0, 10).map((n) => {
            const d = n.created_at ? new Date(n.created_at) : null
            const dateStr = d
              ? `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
              : ''
            return (
              <div key={n.id} className="sn-worknote-card">
                <div className="sn-worknote-header">
                  <span className="sn-worknote-user">{users.find((u) => u.id === n.created_by)?.display_name ?? n.created_by ?? '—'}</span>
                  <span className="sn-worknote-datetime">Work Notes · {dateStr}</span>
                </div>
                <div className="sn-worknote-body">{n.content}</div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

function RelatedListObservables({ caseId, detail }: { caseId: string; detail: CaseDetail; onUpdated: () => void }) {
  const observables = detail.observables ?? []

  return (
    <div className="sn-related-list-panel">
      <div className="sn-related-list-filter" style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
        <Link to="/observables/new" state={{ caseId }} className="sn-btn-new">
          New
        </Link>
        <Link to={`/cases/${caseId}/observables/edit`} className="sn-btn-new">
          Edit
        </Link>
        <span style={{ color: 'var(--sn-text-secondary)', fontSize: '0.9rem' }}>
          New: create an observable. Edit: search and add or remove observables for this case.
        </span>
      </div>
      <table>
        <thead>
          <tr>
            <th>Observable</th>
            <th>Observable Type</th>
            <th>Incident count</th>
            <th>Finding</th>
            <th>Updated</th>
          </tr>
        </thead>
        <tbody>
          {observables.length === 0 ? (
            <tr><td colSpan={5} className="sn-empty">No records to display</td></tr>
          ) : (
            observables.map((o) => {
              const updated = o.updated_at || o.created_at
              return (
                <tr key={o.id}>
                  <td><Link to={`/observables/${o.id}`}>{o.value || '—'}</Link></td>
                  <td>{o.type || '—'}</td>
                  <td>{o.incident_count != null ? o.incident_count : '—'}</td>
                  <td>{o.finding ?? '—'}</td>
                  <td>{updated ? new Date(updated).toLocaleString() : '—'}</td>
                </tr>
              )
            })
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{observables.length} to {observables.length} of {observables.length}</div>
    </div>
  )
}

function formatSimilarObservableValues(raw: string | undefined): string {
  if (!raw || !raw.trim()) return '—'
  const t = raw.trim()
  if (t.startsWith('[')) {
    try {
      const arr = JSON.parse(t) as unknown[]
      if (Array.isArray(arr)) return arr.map((v) => String(v)).join(', ') || '—'
    } catch {
      /* ignore */
    }
  }
  return t
}

function RelatedListSimilarIncidents({ detail }: { detail: CaseDetail }) {
  const list = detail.similar_incidents ?? []
  return (
    <div className="sn-related-list-panel">
      <table>
        <thead>
          <tr>
            <th>Task</th>
            <th>Short Description</th>
            <th>Observable</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {list.length === 0 ? (
            <tr><td colSpan={4} className="sn-empty">No similar security incidents</td></tr>
          ) : (
            list.map((s) => (
              <tr key={s.id}>
                <td>
                  {s.similar_case_id ? (
                    <Link to={`/cases/${encodeURIComponent(s.similar_case_id)}`}>
                      {s.similar_case_number || s.similar_case_id}
                    </Link>
                  ) : (
                    s.similar_case_number || '—'
                  )}
                </td>
                <td>{s.similar_case_title ?? '—'}</td>
                <td>{formatSimilarObservableValues(s.shared_observable_values)}</td>
                <td>
                  {s.similar_case_created_at
                    ? new Date(s.similar_case_created_at).toLocaleString()
                    : '—'}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{list.length} to {list.length} of {list.length}</div>
    </div>
  )
}

function RelatedListWorknotes({ caseId, detail, onUpdated }: { caseId: string; detail: CaseDetail; onUpdated: () => void }) {
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const worknotes = detail.worknotes ?? []

  const handlePost = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return
    setSubmitting(true)
    setError(null)
    try {
      await api.addWorknote(caseId, content.trim(), USER_ID)
      setContent('')
      onUpdated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to post work note')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="sn-related-list-panel">
      <div className="sn-related-list-filter">
        <label style={{ marginRight: '0.5rem' }}>Add work note:</label>
        <form onSubmit={handlePost} style={{ display: 'inline-flex', gap: '0.5rem', alignItems: 'center', flex: 1 }}>
          {error && <span className="sn-form-error" style={{ marginRight: '0.5rem' }}>{error}</span>}
          <textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder="Content…" rows={1} style={{ flex: 1, minWidth: 200 }} disabled={submitting} />
          <button type="submit" className="sn-btn-new" disabled={submitting || !content.trim()}>Post</button>
        </form>
      </div>
      <table>
        <thead>
          <tr>
            <th>Content</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {worknotes.length === 0 ? (
            <tr><td colSpan={2} className="sn-empty">No records to display</td></tr>
          ) : (
            worknotes.map((n) => (
              <tr key={n.id}>
                <td style={{ whiteSpace: 'pre-wrap' }}>{n.content}</td>
                <td>{n.created_at ? new Date(n.created_at).toLocaleString() : '—'}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{worknotes.length} to {worknotes.length} of {worknotes.length}</div>
    </div>
  )
}

function RelatedListAttachments({ caseId, detail, onUpdated }: { caseId: string; detail: CaseDetail; onUpdated: () => void }) {
  const [fileName, setFileName] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const attachments = detail.attachments ?? []

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!fileName.trim()) return
    setSubmitting(true)
    try {
      await api.createAttachment({ case_id: caseId, file_name: fileName.trim() }, USER_ID)
      setFileName('')
      onUpdated()
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="sn-related-list-panel">
      <div className="sn-related-list-filter">
        <label style={{ marginRight: '0.5rem' }}>Add attachment (metadata):</label>
        <form onSubmit={handleAdd} style={{ display: 'inline-flex', gap: '0.5rem', alignItems: 'center' }}>
          <input type="text" value={fileName} onChange={(e) => setFileName(e.target.value)} placeholder="File name" style={{ width: 200, padding: '0.3rem 0.5rem' }} disabled={submitting} />
          <button type="submit" className="sn-btn-new" disabled={submitting || !fileName.trim()}>{submitting ? 'Adding…' : 'Add attachment'}</button>
        </form>
      </div>
      <table>
        <thead>
          <tr>
            <th>File name</th>
            <th>Size</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {attachments.length === 0 ? (
            <tr><td colSpan={3} className="sn-empty">No records to display</td></tr>
          ) : (
            attachments.map((a) => (
              <tr key={a.id}>
                <td>{a.file_name}</td>
                <td>{a.size_bytes != null ? `${a.size_bytes} B` : '—'}</td>
                <td>{a.created_at ? new Date(a.created_at).toLocaleString() : '—'}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{attachments.length} to {attachments.length} of {attachments.length}</div>
    </div>
  )
}

function RelatedListAlerts({ detail }: { detail: CaseDetail }) {
  const alerts = detail.alerts ?? []
  return (
    <div className="sn-related-list-panel">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Summary</th>
            <th>Severity</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {alerts.length === 0 ? (
            <tr><td colSpan={4} className="sn-empty">No records to display</td></tr>
          ) : (
            alerts.map((a) => (
              <tr key={a.id}>
                <td>{a.id}</td>
                <td>{a.summary ?? '—'}</td>
                <td>{a.severity ?? '—'}</td>
                <td>{a.created_at ? new Date(a.created_at).toLocaleString() : '—'}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{alerts.length} to {alerts.length} of {alerts.length}</div>
    </div>
  )
}

function RelatedListEnrichment({
  caseId,
  detail,
}: {
  caseId: string
  detail: CaseDetail
}) {
  const list = detail.enrichment_results ?? []
  const observables = detail.observables ?? []
  const observableById = new Map(observables.map((o) => [o.id, o]))

  return (
    <div className="sn-related-list-panel">
      <div className="sn-related-list-filter" style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
        <Link to={`/cases/${caseId}/enrichment/new`} className="sn-btn-new">
          New
        </Link>
        <span style={{ color: 'var(--sn-text-secondary)', fontSize: '0.9rem' }}>
          New: create an enrichment result for this case.
        </span>
      </div>

      <table>
        <thead>
          <tr>
            <th>Observable</th>
            <th>Source</th>
            <th>Result</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {list.length === 0 ? (
            <tr><td colSpan={4} className="sn-empty">No records to display</td></tr>
          ) : (
            list.map((r) => {
              const obs = observableById.get(r.observable_id)
              const displayValue = obs?.value || r.observable_id || '—'
              return (
                <tr key={r.id}>
                  <td>{displayValue}</td>
                  <td>{r.source ?? '—'}</td>
                  <td>{r.result ?? '—'}</td>
                  <td>{r.created_at ? new Date(r.created_at).toLocaleString() : '—'}</td>
                </tr>
              )
            })
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{list.length} to {list.length} of {list.length}</div>
    </div>
  )
}

function RelatedListAudit({ detail }: { detail: CaseDetail }) {
  const events = detail.audit_events ?? []
  return (
    <div className="sn-related-list-panel">
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>Event</th>
            <th>Action</th>
            <th>Entity</th>
            <th>Actor</th>
            <th>Summary</th>
          </tr>
        </thead>
        <tbody>
          {events.length === 0 ? (
            <tr><td colSpan={6} className="sn-empty">No records to display. Start the case service with AUDIT_SERVICE_ADDR=localhost:50056 and the audit service on 50056 so changes are recorded.</td></tr>
          ) : (
            events.map((e) => (
              <tr key={e.event_id}>
                <td>{e.occurred_at ? new Date(e.occurred_at).toLocaleString() : '—'}</td>
                <td>{e.event_type || '—'}</td>
                <td>{e.action}</td>
                <td>{e.entity_type} {e.entity_id}</td>
                <td>{e.actor_name || e.actor_user_id || '—'}</td>
                <td>{e.change_summary || '—'}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <div className="sn-pagination">{events.length} to {events.length} of {events.length}</div>
    </div>
  )
}
