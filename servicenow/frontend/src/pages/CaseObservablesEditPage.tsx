import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ServicenowLayout } from '../components/ServicenowLayout'
import { api, type Observable } from '../api/client'

const USER_ID = 'frontend-user'

export function CaseObservablesEditPage() {
  const { caseId } = useParams<{ caseId: string }>()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [searchDebounced, setSearchDebounced] = useState('')
  const [collection, setCollection] = useState<Observable[]>([])
  const [associated, setAssociated] = useState<Observable[]>([])
  const [selectedLeft, setSelectedLeft] = useState<Set<string>>(new Set())
  const [selectedRight, setSelectedRight] = useState<Set<string>>(new Set())
  const [loadingLeft, setLoadingLeft] = useState(false)
  const [loadingRight, setLoadingRight] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadCollection = useCallback(async () => {
    if (!caseId) return
    setLoadingLeft(true)
    setError(null)
    try {
      const r = await api.listAllObservables(USER_ID, searchDebounced, 100, '')
      setCollection(r.observables ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load observables')
    } finally {
      setLoadingLeft(false)
    }
  }, [caseId, searchDebounced])

  const loadAssociated = useCallback(async () => {
    if (!caseId) return
    setLoadingRight(true)
    try {
      const r = await api.listObservables(caseId, USER_ID, 200, '')
      setAssociated(r.observables ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load associated observables')
    } finally {
      setLoadingRight(false)
    }
  }, [caseId])

  useEffect(() => {
    const t = setTimeout(() => setSearchDebounced(search), 300)
    return () => clearTimeout(t)
  }, [search])

  useEffect(() => { loadCollection() }, [loadCollection])
  useEffect(() => { loadAssociated() }, [loadAssociated])

  const addToCase = async () => {
    if (!caseId || selectedLeft.size === 0) return
    setBusy(true)
    setError(null)
    try {
      for (const id of selectedLeft) {
        await api.linkObservable(caseId, id, USER_ID)
      }
      setSelectedLeft(new Set())
      await loadAssociated()
      await loadCollection()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to link')
    } finally {
      setBusy(false)
    }
  }

  const removeFromCase = async () => {
    if (!caseId || selectedRight.size === 0) return
    setBusy(true)
    setError(null)
    try {
      for (const id of selectedRight) {
        await api.unlinkObservableFromCase(caseId, id, USER_ID)
      }
      setSelectedRight(new Set())
      await loadAssociated()
      await loadCollection()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to unlink')
    } finally {
      setBusy(false)
    }
  }

  const toggleLeft = (id: string) => {
    setSelectedLeft((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleRight = (id: string) => {
    setSelectedRight((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const associatedIds = new Set(associated.map((o) => o.id))
  const collectionFiltered = collection.filter((o) => !associatedIds.has(o.id))

  return (
    <ServicenowLayout contextTitle={`Edit observables · Case ${caseId ?? ''}`}>
      <main className="app-main sn-observables-edit-page">
        <div className="sn-page sn-observables-edit-page-inner">
          <h2 style={{ margin: 0, fontSize: '1.1rem', marginBottom: '1rem' }}>Edit observables for case</h2>
          {error && (
            <div style={{ padding: '0.5rem 0.75rem', marginBottom: '1rem', background: 'var(--sn-bg-danger-subtle)', color: 'var(--sn-text-danger)', borderRadius: 4 }}>
              {error}
            </div>
          )}
          <div className="sn-observables-edit">
            <div className="sn-observables-edit-search-row">
              <label className="sn-observables-edit-search-label">Search</label>
              <input
                type="text"
                className="sn-observables-edit-search"
                placeholder="Search observables…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <div className="sn-observables-edit-row">
              <div className="sn-observables-box">
                <div className="sn-observables-box-header">Collection</div>
                <div className="sn-observables-box-list">
                  {loadingLeft ? (
                    <div className="sn-observables-box-empty">Loading…</div>
                  ) : collectionFiltered.length === 0 ? (
                    <div className="sn-observables-box-empty">No observables to show</div>
                  ) : (
                    <ul className="sn-observables-list">
                      {collectionFiltered.map((o) => (
                        <li
                          key={o.id}
                          className={`sn-observables-list-item ${selectedLeft.has(o.id) ? 'selected' : ''}`}
                          onClick={() => toggleLeft(o.id)}
                        >
                          <span className="sn-observables-item-value">{o.value || o.id}</span>
                          {o.type && <span className="sn-observables-item-type">({o.type})</span>}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
              <div className="sn-observables-transfer">
                <button
                  type="button"
                  className="sn-btn sn-btn-secondary sn-observables-transfer-btn"
                  disabled={selectedLeft.size === 0 || busy}
                  onClick={addToCase}
                  title="Add selected to case"
                >
                  &gt;
                </button>
                <button
                  type="button"
                  className="sn-btn sn-btn-secondary sn-observables-transfer-btn"
                  disabled={selectedRight.size === 0 || busy}
                  onClick={removeFromCase}
                  title="Remove selected from case"
                >
                  &lt;
                </button>
              </div>
              <div className="sn-observables-box">
                <div className="sn-observables-box-header">Associated to case</div>
                <div className="sn-observables-box-list">
                  {loadingRight ? (
                    <div className="sn-observables-box-empty">Loading…</div>
                  ) : associated.length === 0 ? (
                    <div className="sn-observables-box-empty">No observables associated</div>
                  ) : (
                    <ul className="sn-observables-list">
                      {associated.map((o) => (
                        <li
                          key={o.id}
                          className={`sn-observables-list-item ${selectedRight.has(o.id) ? 'selected' : ''}`}
                          onClick={() => toggleRight(o.id)}
                        >
                          <span className="sn-observables-item-value">{o.value || o.id}</span>
                          {o.type && <span className="sn-observables-item-type">({o.type})</span>}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </div>
          </div>
          <div className="sn-observables-edit-actions">
            <button type="button" className="sn-btn sn-btn-secondary sn-observables-edit-btn" onClick={() => navigate(`/cases/${caseId}`)}>
              Cancel
            </button>
            <button type="button" className="sn-btn sn-btn-primary sn-observables-edit-btn" onClick={() => navigate(`/cases/${caseId}`)}>
              Done
            </button>
          </div>
        </div>
      </main>
    </ServicenowLayout>
  )
}
