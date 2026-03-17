import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Case } from '../api/client'

const USER_ID = 'frontend-user'

export function CaseList() {
  const [cases, setCases] = useState<Case[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setError(null)
    api
      .listCases(USER_ID)
      .then((r) => setCases(r.cases ?? []))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load cases'))
      .finally(() => setLoading(false))
  }, [])

  const formatDate = (s?: string) => {
    if (!s) return '—'
    try {
      const d = new Date(s)
      return isNaN(d.getTime()) ? s : d.toLocaleString()
    } catch {
      return s
    }
  }

  return (
    <div className="sn-page sn-content-proportional">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1rem', flexWrap: 'wrap' }}>
        <h2 style={{ margin: 0 }}>Security Incidents</h2>
        <Link to="/cases/new" className="sn-btn sn-btn-primary" style={{ fontSize: 13 }}>
          Create case
        </Link>
      </div>

      {error && (
        <div className="sn-form-error" style={{ marginBottom: '1rem' }}>
          {error}
        </div>
      )}

      {loading ? (
        <p style={{ color: '#666' }}>Loading cases…</p>
      ) : (
        <div className="sn-table-wrapper" style={{ background: '#fff', border: '1px solid var(--sn-border)', borderRadius: 4 }}>
          <table style={{ width: '100%', minWidth: 600, borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: 'var(--sn-bg-secondary)', borderBottom: '1px solid var(--sn-border)' }}>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}>Number</th>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}>Title</th>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}>State</th>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}>Priority</th>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}>Opened</th>
                <th style={{ textAlign: 'left', padding: '0.5rem 0.75rem', fontWeight: 600 }}></th>
              </tr>
            </thead>
            <tbody>
              {cases.length === 0 ? (
                <tr>
                  <td colSpan={6} style={{ padding: '1.5rem', color: '#666', textAlign: 'center' }}>
                    No cases found.
                  </td>
                </tr>
              ) : (
                cases.map((c) => (
                  <tr key={c.id} style={{ borderBottom: '1px solid var(--sn-border)' }}>
                    <td style={{ padding: '0.5rem 0.75rem' }}>{c.case_number ?? c.id}</td>
                    <td style={{ padding: '0.5rem 0.75rem' }}>{c.title || '—'}</td>
                    <td style={{ padding: '0.5rem 0.75rem' }}>{c.state ?? '—'}</td>
                    <td style={{ padding: '0.5rem 0.75rem' }}>{c.priority ?? '—'}</td>
                    <td style={{ padding: '0.5rem 0.75rem' }}>{formatDate(c.created_at)}</td>
                    <td style={{ padding: '0.5rem 0.75rem' }}>
                      <Link to={`/cases/${encodeURIComponent(c.id)}`} className="sn-btn sn-btn-primary" style={{ fontSize: 13 }}>
                        Open
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
