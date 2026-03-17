import { useState, useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { ServicenowLayout } from './components/ServicenowLayout'
import { CaseList } from './pages/CaseList'
import { CaseForm } from './pages/CaseForm'
import { NewCaseForm } from './pages/NewCaseForm'
import { ObservableNewPage } from './pages/ObservableNewPage'
import { ObservableDetailPage } from './pages/ObservableDetailPage'
import { CaseObservablesEditPage } from './pages/CaseObservablesEditPage'
import { EnrichmentNewPage } from './pages/EnrichmentNewPage'
import { Home } from './pages/Home'
import './App.css'

const GATE_PASSWORD = 'power-raingers'
const GATE_KEY = 'app_unlocked'

function App() {
  const [unlocked, setUnlocked] = useState(() => sessionStorage.getItem(GATE_KEY) === '1')

  useEffect(() => {
    if (sessionStorage.getItem(GATE_KEY) === '1') {
      setUnlocked(true)
      return
    }
    const p = window.prompt('Password')
    if (p === GATE_PASSWORD) {
      sessionStorage.setItem(GATE_KEY, '1')
      setUnlocked(true)
    }
  }, [])

  if (!unlocked) return <div className="app" style={{ padding: '2rem', textAlign: 'center' }}>Access denied.</div>

  return (
    <div className="app">
      <Routes>
        <Route
          path="/"
          element={
            <ServicenowLayout contextTitle="Home">
              <main className="app-main">
                <Home />
              </main>
            </ServicenowLayout>
          }
        />
        <Route
          path="/cases"
          element={
            <ServicenowLayout contextTitle="Security Incidents">
              <main className="app-main">
                <CaseList />
              </main>
            </ServicenowLayout>
          }
        />
        <Route path="/cases/new" element={<NewCaseForm />} />
        <Route path="/observables/new" element={<ObservableNewPage />} />
        <Route path="/observables/:id" element={<ObservableDetailPage />} />
        <Route path="/cases/:caseId/observables/edit" element={<CaseObservablesEditPage />} />
        <Route path="/cases/:caseId/enrichment/new" element={<EnrichmentNewPage />} />
        <Route path="/cases/:caseId" element={<CaseForm />} />
      </Routes>
    </div>
  )
}

export default App
