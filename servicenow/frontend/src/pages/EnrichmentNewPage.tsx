import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api, type Observable } from '../api/client'

const USER_ID = 'frontend-user'

export function EnrichmentNewPage() {
  const { caseId } = useParams<{ caseId: string }>()
  const navigate = useNavigate()

  const [observables, setObservables] = useState<Observable[]>([])
  const [observableId, setObservableId] = useState('')
  const [sourceName, setSourceName] = useState('')
  const [enrichmentType, setEnrichmentType] = useState('')
  const [summary, setSummary] = useState('')
  const [resultData, setResultData] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [caseNumber, setCaseNumber] = useState<string | null>(null)

  // Load observables and case detail for this case.
  useEffect(() => {
    if (!caseId) return
    let cancelled = false
    const load = async () => {
      try {
        const [obsRes, detail] = await Promise.all([
          api.listObservables(caseId, USER_ID, 200, ''),
          api.getCaseDetail(caseId, USER_ID),
        ])
        if (!cancelled) {
          setObservables(obsRes.observables ?? [])
          setCaseNumber(detail.case.case_number || detail.case.id)
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Failed to load observables for case')
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [caseId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!caseId) {
      setError('Case ID is missing from URL.')
      return
    }
    if (!observableId || !sourceName.trim() || !enrichmentType.trim() || !resultData.trim()) {
      setError('Observable, Source, Type, and Result are required.')
      return
    }

    setSubmitting(true)
    setError(null)
    setSuccessMessage(null)

    try {
      let parsedResult: any = {}
      try {
        parsedResult = JSON.parse(resultData)
      } catch {
        parsedResult = { value: resultData.trim() }
      }

      const res = await fetch('/api/v1/enrichment-results', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-Id': USER_ID,
        },
        body: JSON.stringify({
          id: '',
          case_id: caseId,
          observable_id: observableId,
          enrichment_type: enrichmentType.trim(),
          source_name: sourceName.trim(),
          status: 'completed',
          summary: summary.trim() || undefined,
          result_data: parsedResult,
          received_at: new Date().toISOString(),
        }),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || `HTTP ${res.status}`)
      }

      setSuccessMessage('Enrichment result created.')
      setObservableId('')
      setSourceName('')
      setEnrichmentType('')
      setSummary('')
      setResultData('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create enrichment result')
    } finally {
      setSubmitting(false)
    }
  }

  const backHref = caseId ? `/cases/${caseId}` : '/cases'

  return (
    <ServicenowLayout contextTitle="New enrichment result">
      <main className="app-main">
        <div className="sn-form-proportional">
          <h1 className="sn-page-title">New Record | Enrichment Result</h1>
          {caseId && (
            <p style={{ marginBottom: '0.75rem', color: 'var(--sn-text-secondary)' }}>
              Case: <strong>{caseNumber ?? caseId}</strong>
            </p>
          )}
          <p style={{ marginBottom: '1rem', color: 'var(--sn-text-secondary)' }}>
            Create an enrichment result for a specific observable in this case.
          </p>

          <form onSubmit={handleSubmit}>
            <div className="sn-form-group">
              <label htmlFor="observableId">* Observable</label>
              <select
                id="observableId"
                value={observableId}
                onChange={(e) => setObservableId(e.target.value)}
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting || observables.length === 0}
              >
                <option value="">— Select observable —</option>
                {observables.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.value || o.id} {o.type ? `(${o.type})` : ''}
                  </option>
                ))}
              </select>
              {observables.length === 0 && (
                <p style={{ marginTop: '0.25rem', fontSize: 12, color: 'var(--sn-text-secondary)' }}>
                  No observables linked to this case. Add observables first, then create enrichment results.
                </p>
              )}
            </div>

            <div className="sn-form-group">
              <label htmlFor="sourceName">* Source</label>
              <input
                id="sourceName"
                type="text"
                value={sourceName}
                onChange={(e) => setSourceName(e.target.value)}
                placeholder="e.g. VirusTotal, Internal Tool"
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            <div className="sn-form-group">
              <label htmlFor="enrichmentType">* Enrichment type</label>
              <input
                id="enrichmentType"
                type="text"
                value={enrichmentType}
                onChange={(e) => setEnrichmentType(e.target.value)}
                placeholder="e.g. reputation, whois, sandbox"
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            <div className="sn-form-group">
              <label htmlFor="summary">Summary</label>
              <input
                id="summary"
                type="text"
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                placeholder="Short human-readable summary"
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            <div className="sn-form-group">
              <label htmlFor="resultData">* Result data (JSON or text)</label>
              <textarea
                id="resultData"
                value={resultData}
                onChange={(e) => setResultData(e.target.value)}
                rows={5}
                placeholder='e.g. {"verdict":"malicious","score":95} or a plain text description'
                style={{ width: '100%', padding: '0.5rem' }}
                disabled={submitting}
              />
            </div>

            {error && (
              <div className="sn-form-error" style={{ marginBottom: '1rem' }}>
                {error}
              </div>
            )}
            {successMessage && (
              <div
                className="sn-form-success"
                style={{
                  marginBottom: '1rem',
                  padding: '0.75rem',
                  background: 'var(--sn-bg-success, #e8f5e9)',
                  borderRadius: 4,
                }}
              >
                {successMessage}
              </div>
            )}

            <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
              <button type="submit" className="sn-btn-primary" disabled={submitting}>
                {submitting ? 'Saving…' : 'Save'}
              </button>
              <button
                type="button"
                className="sn-btn-secondary"
                onClick={() => navigate(backHref)}
                disabled={submitting}
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </main>
    </ServicenowLayout>
  )
}

