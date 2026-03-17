import { test, expect } from '@playwright/test'

/**
 * Assignment group dropdown: data comes from gateway -> assignment-reference-service -> DB.
 * Requires: gateway (port 8080) and assignment-reference-service (gRPC 50054) running,
 * and assignment_ref DB seeded (run migrations or service self-seeds on startup).
 */
test.describe('Assignment group dropdown', () => {
  test('New Case form shows assignment group options when API returns groups', async ({ page }) => {
    // Wait for reference/groups API to complete
    const groupsResponse = page.waitForResponse((r) =>
      r.url().includes('/api/v1/reference/groups') && r.request().method() === 'GET',
      { timeout: 15000 }
    )
    await page.goto('/cases/new')
    await groupsResponse

    const label = page.getByText('Assignment Group', { exact: false }).first()
    await expect(label).toBeVisible()
    const select = page.locator('select').filter({ has: page.locator('option[value=""]') }).first()
    await expect(select).toBeVisible()

    const errorEl = page.locator('.sn-form-error').filter({ hasText: /assignment|group|load|fail/i })
    if ((await errorEl.count()) > 0) {
      const errText = await errorEl.first().textContent()
      throw new Error(`Assignment groups failed to load: ${errText?.trim() ?? 'unknown'}`)
    }

    const options = select.locator('option')
    const count = await options.count()
    expect(count, 'Assignment group dropdown should have at least one option besides "— None —"').toBeGreaterThan(1)
  })

  test('GET /api/v1/reference/groups returns non-empty groups when backend is seeded', async ({ request }) => {
    // Hit gateway (via frontend proxy or directly)
    const baseURL = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173'
    const apiBase = baseURL.includes('5173') ? 'http://localhost:8080' : baseURL.replace('5173', '8080')
    const res = await request.get(`${apiBase}/api/v1/reference/groups?page_size=10`, {
      headers: { 'X-User-Id': 'e2e-test' },
    })
    expect(res.ok(), `Groups API should return 200, got ${res.status()}`).toBe(true)
    const body = await res.json()
    const groups = body?.groups ?? []
    expect(Array.isArray(groups)).toBe(true)
    expect(groups.length, 'Gateway should return at least one assignment group when DB is seeded').toBeGreaterThan(0)
    expect(groups[0]).toHaveProperty('id')
    expect(groups[0]).toHaveProperty('name')
  })
})
