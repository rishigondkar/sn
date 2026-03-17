import { useState } from 'react'
import { useNavigate, useLocation, Link } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api } from '../api/client'
import type { Observable } from '../api/client'

const USER_ID = 'frontend-user'

// Observable types: value (stored/API) and label (display). Order matches classification precedence where relevant.
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

export function ObservableNewPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const caseIdFromState = (location.state as { caseId?: string } | null)?.caseId

  const [value, setValue] = useState('')
  const [observableType, setObservableType] = useState<string>('ipv4')
  const [notes, setNotes] = useState('')
  const [classifying, setClassifying] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [created, setCreated] = useState<Observable | null>(null)

  const handleClassify = async () => {
    if (!value.trim()) return
    setClassifying(true)
    setError(null)
    try {
      const r = await api.classifyObservable(value.trim(), USER_ID)
      if (r.observable_type) setObservableType(r.observable_type)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Classification failed')
    } finally {
      setClassifying(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!value.trim()) {
      setError('Value is required')
      return
    }
    setSubmitting(true)
    setError(null)
    setCreated(null)
    try {
      const obs = await api.createStandaloneObservable(observableType, value.trim(), USER_ID)
      setCreated(obs)
      setValue('')
      setNotes('')
      if (caseIdFromState) {
        // Optionally auto-link to case and go back
        try {
          await api.linkObservable(caseIdFromState, obs.id, USER_ID)
          navigate(`/cases/${caseIdFromState}`, { state: null })
        } catch {
          // Leave user on page; they can link manually
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create observable')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <ServicenowLayout contextTitle="New observable">
      <main className="app-main">
        <div className="sn-form-proportional">
          <h1 className="sn-page-title">New Record | Observable</h1>
          <p style={{ marginBottom: '1rem', color: 'var(--sn-text-secondary)' }}>
            Enter a value and use <strong>Classify</strong> to auto-detect the observable type, or choose the type manually.
          </p>

          <form onSubmit={handleSubmit}>
            <div className="sn-form-group">
              <label htmlFor="value">* Value</label>
              <input
                id="value"
                type="text"
                value={value}
                onChange={(e) => setValue(e.target.value)}
                placeholder="e.g. 192.168.1.1, example.com, user@example.com"
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            <div className="sn-form-group" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
              <label style={{ marginRight: '0.25rem' }}>Observable type</label>
              <select
                value={observableType}
                onChange={(e) => setObservableType(e.target.value)}
                style={{ padding: '0.4rem 0.6rem', minWidth: 280 }}
                disabled={submitting}
              >
                {OBSERVABLE_TYPE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
              <button
                type="button"
                onClick={handleClassify}
                disabled={classifying || !value.trim()}
                className="sn-btn-secondary"
              >
                {classifying ? 'Classifying…' : 'Classify'}
              </button>
            </div>

            <div className="sn-form-group">
              <label htmlFor="notes">Notes</label>
              <textarea
                id="notes"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            {error && (
              <div className="sn-form-error" style={{ marginBottom: '1rem' }}>
                {error}
              </div>
            )}

            {created && (
              <div className="sn-form-success" style={{ marginBottom: '1rem', padding: '0.75rem', background: 'var(--sn-bg-success, #e8f5e9)', borderRadius: 4 }}>
                Observable created: <strong>{created.id}</strong> ({created.type}: {created.value})
                {caseIdFromState && <span> — linked to case and redirecting…</span>}
                {!caseIdFromState && (
                  <span> — <Link to="/cases">Open a case</Link> to link this observable.</span>
                )}
              </div>
            )}

            <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
              <button type="submit" className="sn-btn-primary" disabled={submitting || !value.trim()}>
                {submitting ? 'Submitting…' : 'Submit'}
              </button>
              <Link to={caseIdFromState ? `/cases/${caseIdFromState}` : '/cases'} className="sn-btn-secondary">
                Cancel
              </Link>
            </div>
          </form>
        </div>
      </main>
    </ServicenowLayout>
  )
}
