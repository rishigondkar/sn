# SOC Case Management — Frontend

React + Vite + TypeScript SPA with a **ServiceNow SIR–style UI** that uses all API Gateway BFF endpoints.

## Stack

- **React 18** + **TypeScript**
- **Vite 6** — dev server, HMR, production build
- **React Router 6** — client-side routing

## ServiceNow-style UI

- **Header**: Dark bar with app title and nav (Home, Cases).
- **Case list** (`/cases`): Open case by ID, or create new case (short description + priority).
- **Case form** (`/cases/:caseId`): Single form with sections and related lists matching the backend:
  - **Case**: Number (read-only), state badge, short description, priority — **Update** button.
  - **Description**: Text area — **Update** button.
  - **Assignment**: Assigned to (users), Assignment group (groups) — **Assign** button.
  - **Close case**: Resolution text area — **Close case** button (hidden when state is closed).
  - **Work notes**: Add work note (textarea) — **Add**; table of notes.
  - **Observables**: Link observable (ID input) — **Link observable**; table of observables.
  - **Alerts**: Read-only table.
  - **Enrichment results**: Read-only table.
  - **Attachments**: Add attachment (file name, size, content type) — **Add attachment**; table.
  - **Audit events**: Read-only table.

All forms and buttons call the corresponding gateway REST APIs.

## Setup

```bash
cd frontend
npm install
```

## Development

Start the dev server (default: http://localhost:5173):

```bash
npm run dev
```

The Vite config proxies `/api` and `/health` to the gateway at `http://localhost:8080`. Ensure the API Gateway (and backend services) are running.

## Build & preview

```bash
npm run build
npm run preview
```

## Project layout

```
frontend/
├── index.html
├── package.json
├── vite.config.ts
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── api/
│   │   └── client.ts    # All gateway endpoints
│   ├── index.css        # ServiceNow-style variables & components
│   └── pages/
│       ├── Home.tsx
│       ├── CaseList.tsx  # Open by ID / Create new
│       └── CaseForm.tsx # Full case form + related lists
└── README.md
```

## Routes

- `/` — Home (backend health).
- `/cases` — List: open by case ID or create new case.
- `/cases/:caseId` — Case form with all sections and related lists.

Auth is simulated via `X-User-Id: frontend-user` in the API client.
