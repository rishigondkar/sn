/**
 * API client for the Gateway BFF.
 * In dev, Vite proxies /api and /health to the gateway (see vite.config.ts).
 */

const API_BASE = ''

function headers(userId: string, json = false): HeadersInit {
  const h: Record<string, string> = { 'X-User-Id': userId }
  if (json) h['Content-Type'] = 'application/json'
  return h
}

async function handleRes<T>(res: Response, parseJson = true): Promise<T> {
  if (!res.ok) {
    const err = await res.json().catch(() => ({})) as {
      error?: { message?: string; details?: Array<{ field?: string; issue?: string }> }
    }
    const msg = err?.error?.message || res.statusText
    const details = err?.error?.details
    const detailStr = details?.length
      ? details.map((d) => `${d.field ?? ''}: ${d.issue ?? ''}`).filter(Boolean).join('; ')
      : ''
    throw new Error(detailStr ? `${msg} (${detailStr})` : msg)
  }
  if (!parseJson || res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  async health(): Promise<{ ok: boolean }> {
    const res = await fetch(`${API_BASE}/health`)
    if (!res.ok) throw new Error(`Health check failed: ${res.status}`)
    return { ok: true }
  },

  // ——— Cases ———
  async createCase(
    body: CreateCaseBody,
    userId: string
  ): Promise<Case> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      })
    )
  },

  async getCase(caseId: string, userId: string): Promise<Case> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}`, { headers: headers(userId) })
    )
  },

  async listCases(
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ cases: Case[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases${q}`, { headers: headers(userId) })
    )
  },

  async updateCase(
    caseId: string,
    body: UpdateCaseBody,
    userId: string
  ): Promise<Case> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}`, {
        method: 'PATCH',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      })
    )
  },

  async assignCase(
    caseId: string,
    body: { assigned_user_id?: string; assignment_group_id?: string },
    userId: string
  ): Promise<void> {
    await handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/assign`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      }),
      false
    )
  },

  async closeCase(caseId: string, body: { resolution?: string }, userId: string): Promise<void> {
    await handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/close`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      }),
      false
    )
  },

  // ——— Case detail (aggregated) ———
  async getCaseDetail(caseId: string, userId: string): Promise<CaseDetail> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/detail`, { headers: headers(userId) })
    )
  },

  // ——— Worknotes ———
  async addWorknote(caseId: string, content: string, userId: string): Promise<Worknote> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/worknotes`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify({ content }),
      })
    )
  },

  async listWorknotes(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ worknotes: Worknote[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/worknotes${q}`, {
        headers: headers(userId),
      })
    )
  },

  // ——— Observables ———
  async linkObservable(
    caseId: string,
    observableId: string,
    userId: string
  ): Promise<void> {
    await handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/observables`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify({ observable_id: observableId }),
      }),
      false
    )
  },

  /** Classify a raw value: returns suggested observable_type and normalized_value. */
  async classifyObservable(
    value: string,
    userId: string
  ): Promise<{ observable_type: string; normalized_value: string }> {
    const params = new URLSearchParams({ value: value.trim() })
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables/classify?${params}`, { headers: headers(userId) })
    )
  },

  /** Create a standalone observable (no case link). Returns the created/existing observable. */
  async createStandaloneObservable(
    observableType: string,
    observableValue: string,
    userId: string
  ): Promise<Observable> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify({
          observable_type: observableType,
          observable_value: observableValue,
        }),
      })
    )
  },

  /** Get a single observable by ID (for the observable detail page). */
  async getObservable(observableId: string, userId: string): Promise<Observable> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables/${observableId}`, { headers: headers(userId) })
    )
  },

  /** Update an observable (PATCH). */
  async updateObservable(
    observableId: string,
    body: UpdateObservableBody,
    userId: string
  ): Promise<Observable> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables/${observableId}`, {
        method: 'PATCH',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      })
    )
  },

  async listObservables(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ observables: Observable[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/observables${q}`, { headers: headers(userId) })
    )
  },

  /** List all observables (for Edit members). Optional search, paginated. */
  async listAllObservables(
    userId: string,
    search = '',
    pageSize = 100,
    pageToken = ''
  ): Promise<{ observables: Observable[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (search) params.set('search', search)
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables${q}`, { headers: headers(userId) })
    )
  },

  async unlinkObservableFromCase(caseId: string, observableId: string, userId: string): Promise<void> {
    await handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/observables/${observableId}`, {
        method: 'DELETE',
        headers: headers(userId),
      }),
      false
    )
  },

  async listAlerts(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ alerts: Alert[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/alerts${q}`, { headers: headers(userId) })
    )
  },

  async listEnrichmentResults(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ enrichment_results: EnrichmentResult[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/enrichment-results${q}`, {
        headers: headers(userId),
      })
    )
  },

  async listThreatLookups(
    observableId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ threat_lookups: ThreatLookupResult[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/observables/${observableId}/threat-lookups${q}`, {
        headers: headers(userId),
      })
    )
  },

  // ——— Attachments ———
  async createAttachment(
    body: { case_id: string; file_name: string; size_bytes?: number; content_type?: string },
    userId: string
  ): Promise<Attachment> {
    return handleRes(
      await fetch(`${API_BASE}/api/v1/attachments`, {
        method: 'POST',
        headers: headers(userId, true),
        body: JSON.stringify(body),
      })
    )
  },

  async listAttachments(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ attachments: Attachment[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/attachments${q}`, {
        headers: headers(userId),
      })
    )
  },

  // ——— Audit ———
  async listAuditEvents(
    caseId: string,
    userId: string,
    pageSize = 50,
    pageToken = ''
  ): Promise<{ audit_events: AuditEvent[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/cases/${caseId}/audit-events${q}`, {
        headers: headers(userId),
      })
    )
  },

  // ——— Reference ———
  async listUsers(
    userId: string,
    pageSize = 100,
    pageToken = ''
  ): Promise<{ users: User[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/reference/users${q}`, { headers: headers(userId) })
    )
  },

  async listGroups(
    userId: string,
    pageSize = 100,
    pageToken = ''
  ): Promise<{ groups: Group[]; next_page_token?: string }> {
    const params = new URLSearchParams()
    if (pageSize) params.set('page_size', String(pageSize))
    if (pageToken) params.set('page_token', pageToken)
    const q = params.toString() ? `?${params}` : ''
    return handleRes(
      await fetch(`${API_BASE}/api/v1/reference/groups${q}`, { headers: headers(userId) })
    )
  },
}

// ——— Types (match gateway DTOs) ———
export type Case = {
  id: string
  case_number: string
  title: string
  state: string
  priority: string
  severity?: string
  description?: string
  assigned_user_id?: string
  assignment_group_id?: string
  opened_by_user_id?: string
  affected_user_id?: string
  created_at?: string
  updated_at?: string
  followup_time?: string
  category?: string
  subcategory?: string
  source?: string
  source_tool?: string
  source_tool_feature?: string
  configuration_item?: string
  soc_notes?: string
  next_steps?: string
  csirt_classification?: string
  soc_lead_user_id?: string
  requested_by_user_id?: string
  environment_level?: string
  environment_type?: string
  pdn?: string
  impacted_object?: string
  notification_time?: string
  opened_time?: string
  event_occurred_time?: string
  mttr?: string
  reassignment_count?: number
  assigned_to_count?: number
  is_affected_user_vip?: boolean
  engineering_document?: string
  response_document?: string
  accuracy?: string
  determination?: string
  closure_reason?: string
  closed_by_user_id?: string
  closed_time?: string
}

export type CreateCaseBody = {
  title: string
  description?: string
  priority?: string
  severity?: string
  affected_user_id?: string
  assigned_user_id?: string
  assignment_group_id?: string
  followup_time?: string
  notification_time?: string
  is_affected_user_vip?: boolean
  category?: string
  subcategory?: string
  source?: string
  source_tool?: string
  source_tool_feature?: string
  configuration_item?: string
  soc_notes?: string
  next_steps?: string
  csirt_classification?: string
  soc_lead_user_id?: string
  requested_by_user_id?: string
  environment_level?: string
  environment_type?: string
  pdn?: string
  impacted_object?: string
  mttr?: string
  engineering_document?: string
  response_document?: string
}

export type UpdateCaseBody = {
  title?: string
  description?: string
  state?: string
  priority?: string
  severity?: string
  affected_user_id?: string
  assigned_user_id?: string
  assignment_group_id?: string
  followup_time?: string
  notification_time?: string
  event_occurred_time?: string
  category?: string
  subcategory?: string
  source?: string
  source_tool?: string
  source_tool_feature?: string
  configuration_item?: string
  soc_notes?: string
  next_steps?: string
  csirt_classification?: string
  soc_lead_user_id?: string
  requested_by_user_id?: string
  environment_level?: string
  environment_type?: string
  pdn?: string
  impacted_object?: string
  mttr?: string
  reassignment_count?: number
  assigned_to_count?: number
  is_affected_user_vip?: boolean
  engineering_document?: string
  response_document?: string
  accuracy?: string
  determination?: string
  closure_reason?: string
}

export type Worknote = {
  id: string
  case_id: string
  content: string
  created_by?: string
  created_at: string
}

export type Alert = {
  id: string
  case_id: string
  summary?: string
  severity?: string
  created_at: string
}

export type Observable = {
  id: string
  case_id: string
  type: string
  value: string
  tracking_status?: string
  created_at: string
  updated_at?: string
  finding?: string
  incident_count?: number
  notes?: string
}

export type UpdateObservableBody = {
  observable_value?: string
  observable_type?: string
  finding?: string
  notes?: string
}

export type EnrichmentResult = {
  id: string
  observable_id: string
  case_id: string
  source?: string
  result?: string
  created_at: string
}

export type ThreatLookupResult = {
  id: string
  observable_id: string
  case_id: string
  provider?: string
  verdict?: string
  created_at: string
}

export type Attachment = {
  id: string
  case_id: string
  file_name: string
  size_bytes?: number
  created_at: string
}

export type AuditEvent = {
  event_id: string
  event_type: string
  entity_type: string
  entity_id: string
  action: string
  actor_user_id?: string
  actor_name?: string
  change_summary?: string
  occurred_at: string
}

export type User = {
  id: string
  display_name?: string
  email?: string
}

export type Group = {
  id: string
  name?: string
}

export type SimilarIncident = {
  id: string
  case_id: string
  similar_case_id?: string
  summary?: string
  similarity?: string
  shared_observable_values?: string
  similar_case_number?: string
  similar_case_title?: string
  similar_case_created_at?: string
}

export type CaseDetail = {
  case: Case
  worknotes?: Worknote[]
  alerts?: Alert[]
  observables?: Observable[]
  similar_incidents?: SimilarIncident[]
  enrichment_results?: EnrichmentResult[]
  threat_lookups?: ThreatLookupResult[]
  attachments?: Attachment[]
  audit_events?: AuditEvent[]
  assigned_user?: User
  assignment_group?: Group
  opened_by_user?: User
  degraded_sections?: string[]
}
