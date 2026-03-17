import { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'

interface ServicenowLayoutProps {
  contextTitle: string
  actionBar?: ReactNode
  children: ReactNode
}

export function ServicenowLayout({ contextTitle, actionBar, children }: ServicenowLayoutProps) {
  return (
    <>
      <header className="sn-global-header">
        <nav className="sn-nav-links">
          <NavLink to="/cases">Cases</NavLink>
        </nav>
        <div className="sn-context">
          <button type="button" className="sn-context-btn">
            {contextTitle}
            <span aria-hidden>★</span>
          </button>
        </div>
        <input type="text" className="sn-search" placeholder="Search" />
        <div className="sn-user" title="User">RG</div>
      </header>
      {actionBar && <div className="sn-action-bar">{actionBar}</div>}
      {children}
    </>
  )
}
