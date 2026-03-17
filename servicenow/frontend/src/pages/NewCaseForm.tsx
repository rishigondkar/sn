import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api } from '../api/client'

const USER_ID = 'frontend-user'

export function NewCaseForm() {
  const navigate = useNavigate()
  const [shortDescription, setShortDescription] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState('4 - Low')
  const [severity, setSeverity] = useState('4 - Low')
  const [assignmentGroupId, setAssignmentGroupId] = useState('')
  const [assignedUserId, setAssignedUserId] = useState('')
  const [affectedUserId, setAffectedUserId] = useState('')
  const [followupTime, setFollowupTime] = useState('')
  const [notificationTime, setNotificationTime] = useState('')
  const [isAffectedUserVip, setIsAffectedUserVip] = useState(false)
  const [category, setCategory] = useState('')
  const [subcategory, setSubcategory] = useState('')
  const [source, setSource] = useState('')
  const [sourceTool, setSourceTool] = useState('')
  const [sourceToolFeature, setSourceToolFeature] = useState('')
  const [configurationItem, setConfigurationItem] = useState('')
  const [socNotes, setSocNotes] = useState('')
  const [nextSteps, setNextSteps] = useState('')
  const [csirtClassification, setCsirtClassification] = useState('')
  const [socLeadUserId, setSocLeadUserId] = useState('')
  const [users, setUsers] = useState<{ id: string; display_name?: string }[]>([])
  const [groups, setGroups] = useState<{ id: string; name?: string }[]>([])
  const [usersError, setUsersError] = useState<string | null>(null)
  const [groupsError, setGroupsError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setUsersError(null)
    setGroupsError(null)
    api.listUsers(USER_ID).then((r) => setUsers(r.users || [])).catch((err) => setUsersError(err instanceof Error ? err.message : 'Failed to load users'))
    api.listGroups(USER_ID).then((r) => setGroups(r.groups || [])).catch((err) => setGroupsError(err instanceof Error ? err.message : 'Failed to load assignment groups'))
  }, [])

  const priorityToVal = (p: string) =>
    p.startsWith('P') ? p : { '1 - Critical': 'P1', '2 - High': 'P2', '3 - Medium': 'P3', '4 - Low': 'P4' }[p] || 'P3'
  const severityToVal = (s: string) =>
    ({ '1 - Critical': 'critical', '2 - High': 'high', '3 - Medium': 'medium', '4 - Low': 'low' }[s] || s.toLowerCase())

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!shortDescription.trim()) {
      setError('Short description is required')
      return
    }
    if (!notificationTime.trim()) {
      setError('Notification Time is required when creating an incident')
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const c = await api.createCase(
        {
          title: shortDescription.trim(),
          description: description.trim() || undefined,
          priority: priorityToVal(priority),
          severity: severityToVal(severity),
          affected_user_id: affectedUserId || undefined,
          assigned_user_id: assignedUserId || undefined,
          assignment_group_id: assignmentGroupId || undefined,
          followup_time: followupTime ? new Date(followupTime).toISOString() : undefined,
          notification_time: notificationTime ? new Date(notificationTime).toISOString() : undefined,
          is_affected_user_vip: isAffectedUserVip,
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
        },
        USER_ID
      )
      navigate(`/cases/${c.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create incident')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <ServicenowLayout contextTitle="Security Incident - New Record">
      <div className="sn-incident-layout">
        <div className="sn-sidebar">
          <div className="sn-sidebar-title">Security Incident</div>
          <div className="sn-sidebar-number" style={{ color: '#666', fontWeight: 500 }}>New record</div>
        </div>
        <div className="sn-main">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
            <div>
              <Link to="/cases" style={{ fontSize: 13, marginRight: '0.5rem' }}>← Back</Link>
              <span style={{ marginLeft: '0.5rem', fontWeight: 700, fontSize: '1.1rem' }}>Security Incident New record</span>
            </div>
            <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
              <button type="submit" form="new-sir-form" className="sn-btn-bar" disabled={submitting}>{submitting ? 'Submitting…' : 'Submit'}</button>
              <button type="submit" form="new-sir-form" className="sn-btn-bar" disabled={submitting}>{submitting ? 'Saving…' : 'Save'}</button>
              <button type="button" className="sn-btn-bar" disabled>Request IP Block</button>
            </div>
          </div>

          <div className="sn-workflow-stages">
            <span className="sn-stage active">Draft</span>
            <span className="sn-stage">Analysis</span>
            <span className="sn-stage">Contain</span>
            <span className="sn-stage">Eradicate</span>
            <span className="sn-stage">Recover</span>
            <span className="sn-stage">Review</span>
            <span className="sn-stage">Resolved</span>
            <span className="sn-stage">Closed</span>
          </div>

          <div className="sn-tabs">
            <button type="button" className="sn-tab active">Incident Details</button>
            <button type="button" className="sn-tab">Related Records</button>
            <button type="button" className="sn-tab">MITRE ATT&amp;CK Card</button>
            <button type="button" className="sn-tab">Vendor</button>
          </div>

          <form id="new-sir-form" onSubmit={handleSubmit}>
            <div className="sn-new-record-form">
              <div className="sn-form-col">
                <div className="sn-form-group">
                  <label>State</label>
                  <select disabled style={{ background: '#f5f5f5' }}>
                    <option>Draft</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Followup date and time</label>
                  <input type="datetime-local" value={followupTime} onChange={(e) => setFollowupTime(e.target.value)} disabled={submitting} placeholder=" " />
                </div>
                <div className="sn-form-group">
                  <label>Notification Time <span style={{ color: '#c00' }}>*</span></label>
                  <input type="datetime-local" value={notificationTime} onChange={(e) => setNotificationTime(e.target.value)} disabled={submitting} required placeholder=" " />
                </div>
                <div className="sn-form-group">
                  <label>Is Affected User A VIP <span style={{ color: '#c00' }}>*</span></label>
                  <select value={isAffectedUserVip ? 'Yes' : 'No'} onChange={(e) => setIsAffectedUserVip(e.target.value === 'Yes')} disabled={submitting}>
                    <option value="No">No</option>
                    <option value="Yes">Yes</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Substate</label>
                  <select disabled style={{ background: '#f5f5f5' }}>
                    <option>— None —</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Impacted Object</label>
                  <select disabled style={{ background: '#f5f5f5' }}>
                    <option>— None —</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Source</label>
                  <input type="text" value={source} onChange={(e) => setSource(e.target.value)} placeholder="— None —" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Source Tool</label>
                  <input type="text" value={sourceTool} onChange={(e) => setSourceTool(e.target.value)} placeholder="—" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Source Tool Feature</label>
                  <input type="text" value={sourceToolFeature} onChange={(e) => setSourceToolFeature(e.target.value)} placeholder="—" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Configuration item</label>
                  <input type="text" value={configurationItem} onChange={(e) => setConfigurationItem(e.target.value)} placeholder="—" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Affected user</label>
                  <select value={affectedUserId} onChange={(e) => setAffectedUserId(e.target.value)} disabled={submitting || !!usersError}>
                    <option value="">— None —</option>
                    {users.map((u) => (
                      <option key={u.id} value={u.id}>{u.display_name || u.id}</option>
                    ))}
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Category</label>
                  <input type="text" value={category} onChange={(e) => setCategory(e.target.value)} placeholder="— None —" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Subcategory</label>
                  <input type="text" value={subcategory} onChange={(e) => setSubcategory(e.target.value)} placeholder="— None —" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label><span className="required">*</span> Short Description</label>
                  <input
                    type="text"
                    value={shortDescription}
                    onChange={(e) => setShortDescription(e.target.value)}
                    placeholder="Brief summary of the incident"
                    disabled={submitting}
                    required
                  />
                </div>
                <div className="sn-form-group">
                  <label>Engineering Document</label>
                  <textarea readOnly rows={2} style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Response Document</label>
                  <textarea readOnly rows={2} style={{ background: '#f5f5f5' }} />
                </div>
              </div>
              <div className="sn-form-col">
                <div className="sn-form-group">
                  <label>Number</label>
                  <input type="text" value="(auto)" readOnly style={{ background: '#f5f5f5', color: '#888' }} />
                </div>
                <div className="sn-form-group">
                  <label>Opened By</label>
                  <input type="text" value="—" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Opened</label>
                  <input type="text" value={new Date().toLocaleString()} readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Notification Time</label>
                  <input type="text" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Event Time</label>
                  <input type="text" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>MTTR</label>
                  <input type="text" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Priority</label>
                  <select value={priority} onChange={(e) => setPriority(e.target.value)} disabled={submitting}>
                    <option value="1 - Critical">1 - Critical</option>
                    <option value="2 - High">2 - High</option>
                    <option value="3 - Medium">3 - Medium</option>
                    <option value="4 - Low">4 - Low</option>
                    <option value="P1">P1</option>
                    <option value="P2">P2</option>
                    <option value="P3">P3</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Severity</label>
                  <select value={severity} onChange={(e) => setSeverity(e.target.value)} disabled={submitting}>
                    <option value="1 - Critical">1 - Critical</option>
                    <option value="2 - High">2 - High</option>
                    <option value="3 - Medium">3 - Medium</option>
                    <option value="4 - Low">4 - Low</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label><span className="required">*</span> Assignment Group</label>
                  {groupsError && <div className="sn-form-error" style={{ marginBottom: 4 }}>{groupsError}</div>}
                  <select value={assignmentGroupId} onChange={(e) => setAssignmentGroupId(e.target.value)} disabled={submitting || !!groupsError}>
                    <option value="">— None —</option>
                    {groups.map((g) => (
                      <option key={g.id} value={g.id}>{g.name || g.id}</option>
                    ))}
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Assigned To</label>
                  {usersError && <div className="sn-form-error" style={{ marginBottom: 4 }}>{usersError}</div>}
                  <select value={assignedUserId} onChange={(e) => setAssignedUserId(e.target.value)} disabled={submitting || !!usersError}>
                    <option value="">— None —</option>
                    {users.map((u) => (
                      <option key={u.id} value={u.id}>{u.display_name || u.id}</option>
                    ))}
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Reassignment Count</label>
                  <input type="text" value="0" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>Assigned to count</label>
                  <input type="text" value="0" readOnly style={{ background: '#f5f5f5' }} />
                </div>
                <div className="sn-form-group">
                  <label>SOC Lead</label>
                  <select value={socLeadUserId} onChange={(e) => setSocLeadUserId(e.target.value)} disabled={submitting || !!usersError}>
                    <option value="">— None —</option>
                    {users.map((u) => (
                      <option key={u.id} value={u.id}>{u.display_name || u.id}</option>
                    ))}
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>CSIRT Classification</label>
                  <input type="text" value={csirtClassification} onChange={(e) => setCsirtClassification(e.target.value)} placeholder="— None —" disabled={submitting} />
                </div>
                <div className="sn-form-group">
                  <label>Is Affected User A VIP?</label>
                  <select disabled style={{ background: '#f5f5f5' }}>
                    <option>— None —</option>
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Time worked</label>
                  <input type="text" value="00:00:00" readOnly style={{ background: '#f5f5f5', width: '6rem' }} />
                </div>
              </div>
            </div>

            <div style={{ marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid var(--sn-border)' }}>
              <div style={{ fontWeight: 600, marginBottom: '0.5rem' }}>SOC Notes</div>
              <div className="sn-form-group">
                <label>Description</label>
                <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Detailed description" rows={3} disabled={submitting} />
              </div>
              <div className="sn-form-group">
                <label>SOC Notes</label>
                <textarea placeholder="Optional" rows={2} value={socNotes} onChange={(e) => setSocNotes(e.target.value)} disabled={submitting} />
              </div>
              <div className="sn-form-group">
                <label>Next Steps</label>
                <textarea placeholder="Optional" rows={2} value={nextSteps} onChange={(e) => setNextSteps(e.target.value)} disabled={submitting} />
              </div>
              <div className="sn-form-group">
                <label>Work Notes</label>
                <textarea placeholder="Add work notes after creation" rows={2} readOnly style={{ background: '#fef9e6' }} />
              </div>
              <div className="sn-form-group">
                <label>Secure notes</label>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.25rem' }}>
                  <textarea placeholder="—" rows={2} readOnly style={{ flex: 1, background: '#f5f5f5' }} />
                  <span style={{ display: 'flex', flexDirection: 'column' }}>
                    <button type="button" className="sn-btn-small" disabled>−</button>
                    <button type="button" className="sn-btn-small" disabled>+</button>
                  </span>
                </div>
              </div>
            </div>

            <div style={{ marginTop: '1.5rem', display: 'flex', gap: '0.5rem' }}>
              <button type="submit" className="sn-btn sn-btn-primary" disabled={submitting}>{submitting ? 'Submitting…' : 'Submit'}</button>
              <button type="submit" className="sn-btn sn-btn-secondary" disabled={submitting}>Save</button>
              <button type="button" className="sn-btn sn-btn-secondary" disabled>Request IP Block</button>
            </div>
            {error && <p style={{ color: '#c00', marginTop: '0.5rem' }}>{error}</p>}
          </form>
        </div>
      </div>
    </ServicenowLayout>
  )
}
