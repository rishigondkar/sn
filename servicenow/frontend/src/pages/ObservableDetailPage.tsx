import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api } from '../api/client'

const USER_ID = 'frontend-user'

const OBSERVABLE_TYPE_OPTIONS: { value: string; label: string }[] = [
  { value: 'ipv4', label: 'IP Address (V4)' },
  { value: 'ipv6', label: 'IP Address (V6)' },
  { value: 'md5', label: 'MD5 hash' },
  { value: 'sha1', label: 'SHA1 hash' },
  { value: 'sha256', label: 'SHA256 hash' },
  { value: 'sha384', label: 'SHA384 hash' },
  { value: 'sha512', label: 'SHA512 hash' },
  { value: 'file', label: 'File' },
  { value: 'domain', label: 'Domain name' },
  { value: 'hostname', label: 'Host name' },
  { value: 'email', label: 'Email address' },
  { value: 'url', label: 'URL' },
  { value: 'uri', label: 'URI' },
  { value: 'registry_key', label: 'Registry key' },
  { value: 'file_path', label: 'File Path' },
  { value: 'file_name', label: 'File Name' },
  { value: 'pehash', label: 'PEHASH' },
  { value: 'imphash', label: 'IMPHASH' },
  { value: 'cidr', label: 'CIDR rule' },
  { value: 'mutex', label: 'MUTEX name' },
  { value: 'cve', label: 'CVE Number' },
  { value: 'asn', label: 'Autonomous System Number' },
  { value: 'ipv4_network', label: 'IPV4 Network' },
  { value: 'ipv6_network', label: 'IPV6 Network' },
  { value: 'ipv4_netmask', label: 'IPV4 Netmask' },
  { value: 'ipv6_netmask', label: 'IPV6 Netmask' },
  { value: 'tld', label: 'Top-level domain name' },
  { value: 'mac_address', label: 'Mac Address' },
  { value: 'atm_address', label: 'Asynchronous Transfer Mode address' },
  { value: 'windows_executable', label: 'Windows Executable File' },
  { value: 'email_subject', label: 'Email Subject' },
  { value: 'email_message_id', label: 'Email Message ID' },
  { value: 'email_body', label: 'Email Body' },
  { value: 'username', label: 'Username' },
  { value: 'postal_address', label: 'Postal Address' },
  { value: 'certificate_serial', label: 'Certificate Serial Number' },
  { value: 'organization', label: 'Organization name' },
  { value: 'phone_number', label: 'Phone number' },
  { value: 'observable_composition', label: 'Observable Composition' },
  { value: 'command_line', label: 'Command Line' },
  { value: 'ip', label: 'IP (legacy)' },
]

const FINDING_OPTIONS = ['-- None --', 'Unknown', 'Malicious', 'Suspicious', 'Clean']

export function ObservableDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [value, setValue] = useState('')
  const [observableType, setObservableType] = useState('')
  const [finding, setFinding] = useState('')
  const [notes, setNotes] = useState('')
  const [incidentCount, setIncidentCount] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    setError(null)
    api
      .getObservable(id, USER_ID)
      .then((o) => {
        setValue(o.value ?? '')
        setObservableType(o.type ?? '')
        setFinding(o.finding ?? '-- None --')
        setNotes(o.notes ?? '')
        setIncidentCount(o.incident_count ?? null)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load observable'))
      .finally(() => setLoading(false))
  }, [id])

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id) return
    setSaving(true)
    setError(null)
    try {
      await api.updateObservable(
        id,
        {
          observable_value: value,
          observable_type: observableType || undefined,
          finding: finding === '-- None --' ? undefined : finding,
          notes: notes || undefined,
        },
        USER_ID
      )
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <ServicenowLayout contextTitle="Observable">
        <main className="app-main">
          <p>Loading…</p>
        </main>
      </ServicenowLayout>
    )
  }

  if (error && !value && !observableType) {
    return (
      <ServicenowLayout contextTitle="Observable">
        <main className="app-main">
          <p className="sn-form-error">{error}</p>
          <Link to="/cases" style={{ display: 'inline-block', marginTop: '0.5rem' }}>Back to cases</Link>
        </main>
      </ServicenowLayout>
    )
  }

  return (
    <ServicenowLayout contextTitle="Observable">
      <main className="app-main">
        <div className="sn-form-container">
          <h1 className="sn-page-title">Edit Observable</h1>
          <p style={{ margin: 0, fontSize: '0.875rem', color: 'var(--sn-text-secondary, #666)' }}>
            <Link to="/cases">← Back to cases</Link>
          </p>
          <div className="sn-form-card">
            {error && <p className="sn-form-error" style={{ marginBottom: '1rem' }}>{error}</p>}
            <form onSubmit={handleSave}>
              <div className="sn-form-row">
                <div className="sn-form-group">
                  <label>Value</label>
                  <input
                    type="text"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    placeholder="Observable value"
                    disabled={saving}
                  />
                </div>
                <div className="sn-form-group">
                  <label>Incident count</label>
                  <input
                    type="text"
                    value={incidentCount != null ? String(incidentCount) : '—'}
                    readOnly
                    aria-readonly="true"
                  />
                </div>
              </div>
              <div className="sn-form-row">
                <div className="sn-form-group">
                  <label>Observable type</label>
                  <select
                    value={observableType}
                    onChange={(e) => setObservableType(e.target.value)}
                    disabled={saving}
                  >
                    {observableType && !OBSERVABLE_TYPE_OPTIONS.some((o) => o.value === observableType) && (
                      <option value={observableType}>{observableType}</option>
                    )}
                    {OBSERVABLE_TYPE_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="sn-form-group">
                  <label>Finding</label>
                  <select
                    value={finding}
                    onChange={(e) => setFinding(e.target.value)}
                    disabled={saving}
                  >
                    {FINDING_OPTIONS.map((opt) => (
                      <option key={opt} value={opt}>
                        {opt}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="sn-form-row sn-form-row-full">
                <div className="sn-form-group">
                  <label>Notes</label>
                  <textarea
                    value={notes}
                    onChange={(e) => setNotes(e.target.value)}
                    placeholder="Notes (editable)"
                    rows={4}
                    disabled={saving}
                    style={{ minHeight: 96, resize: 'vertical' }}
                  />
                </div>
              </div>
              <div className="sn-form-actions">
                <button type="submit" className="sn-btn sn-btn-primary" disabled={saving}>
                  {saving ? 'Saving…' : 'Save'}
                </button>
                <Link to="/cases" className="sn-btn sn-btn-secondary" style={{ textDecoration: 'none' }}>
                  Cancel
                </Link>
              </div>
            </form>
          </div>
        </div>
      </main>
    </ServicenowLayout>
  )
}
