import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'

export function Home() {
  const [health, setHealth] = useState<'checking' | 'ok' | 'error'>('checking')

  useEffect(() => {
    api
      .health()
      .then(() => setHealth('ok'))
      .catch(() => setHealth('error'))
  }, [])

  return (
    <div className="sn-page sn-form-proportional">
      <h2>Welcome</h2>
      <p>SOC Case Management. Go to <Link to="/cases">Cases</Link> to view all incidents.</p>
      <div style={{ background: '#fff', border: '1px solid var(--sn-border)', marginTop: '1rem' }}>
        <div style={{ padding: '0.5rem 1rem', background: 'var(--sn-table-header-bg)', borderBottom: '1px solid var(--sn-table-border)', fontWeight: 600 }}>System status</div>
        <div style={{ padding: '1rem' }}>Backend: <span style={{ fontWeight: 600 }}>{health}</span></div>
      </div>
    </div>
  )
}
