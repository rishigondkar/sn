# Plan: Make the UI Proportional to Screen Size

**Goal:** Scale layout and content to viewport size without breaking existing behavior. Use a single, consistent approach so the UI feels proportional on different screen sizes.

---

## 1. Principles

- **Use viewport-relative units and caps:** `min(MAX_PX, XXvw)` for width; keep a sensible `max-width` so lines don’t get too long on large monitors.
- **One source of truth:** Prefer global CSS variables and shared classes in `index.css` over one-off inline `style={{ maxWidth: … }}` in components.
- **Progressive change:** Apply changes page-by-page or section-by-section; verify each step before moving on.
- **No breaking layout:** Avoid removing flex/grid or changing structure; only adjust widths, padding, and gaps. Tables stay scrollable where needed.

---

## 2. Current State (Summary)

| Area | Current behavior | Risk if unchanged |
|------|------------------|-------------------|
| **App shell** | `.app-main` full width; header 48px; no max-width on main | Very wide content on large screens |
| **Case form (CaseForm)** | `.sn-incident-layout` flex; sidebar fixed 220px; `.sn-main` flex:1, padding 1rem | Sidebar fixed; main can get very wide |
| **Observable detail** | `.sn-form-container` already `max-width: min(900px, 92vw)` | OK; use as reference pattern |
| **Observable new** | Inline `maxWidth: 640` on container | Override with shared proportional class |
| **Case list** | Inline `maxWidth: 960, margin: '0 auto'` | Move to CSS; make width viewport-relative |
| **New case form** | `.sn-new-record-form` max-width 960px | Make proportional |
| **Home** | Inline `maxWidth: 600` | Make proportional |
| **Tables (related lists)** | `width: 100%` inside panels | Add horizontal scroll on small viewports |
| **Form grids** | `.sn-form-two-cols` max-width 900px | Make proportional |

---

## 3. CSS Variables to Add (in `:root`)

Add to `index.css` so all proportional containers use the same logic:

```css
/* Proportional layout – use in containers */
--sn-content-max: 900px;        /* cap for readability */
--sn-content-width: min(var(--sn-content-max), 92vw);
--sn-content-padding: 1.25rem; /* horizontal padding from viewport edge */
```

Optional for smaller “form-only” pages (e.g. New Observable, Home):

```css
--sn-form-max: 640px;
--sn-form-width: min(var(--sn-form-max), 92vw);
```

---

## 4. Implementation Steps

### Step 1: Add CSS variables and a shared content wrapper class

- In `index.css`:
  - Add the variables above under `:root`.
  - Add a class, e.g. `.sn-content-proportional`, that sets:
    - `width: 100%`
    - `max-width: var(--sn-content-width)` (or `var(--sn-form-width)` for form pages)
    - `margin: 0 auto`
    - `padding-left: var(--sn-content-padding)`; `padding-right: var(--sn-content-padding)`
    - `box-sizing: border-box`
- Do not change any component yet; only add the class where we refactor in later steps.

### Step 2: Main content area (app-main)

- Ensure `.app-main` does not force full bleed on huge screens:
  - Either give `.app-main` a max-width using `var(--sn-content-width)` and center it, or
  - Wrap page content in a single wrapper that uses `.sn-content-proportional` (if layout is always one column).
- Prefer one approach for all “standalone” pages (Case list, Home, New Case, Observable new/detail) so behavior is consistent.

### Step 3: Case form (CaseForm) – sidebar + main

- **Sidebar:** Keep fixed width for now (e.g. 220px). Optionally, on very small viewports (e.g. `< 768px`), hide or collapse sidebar via a media query so main content has room.
- **Main:** Ensure `.sn-main` (or the scrollable content inside it) has a reasonable max-width so form and tables don’t stretch to 100% of a 4K screen. Options:
  - Apply `max-width: var(--sn-content-width)` and `margin: 0 auto` to the inner content wrapper inside `.sn-main`, or
  - Use the same `.sn-content-proportional` wrapper for the form + related lists.
- **Related list panels and tables:** Keep `width: 100%` on the table; add `overflow-x: auto` to the panel (or a wrapper) so on narrow viewports the table scrolls instead of squashing.

### Step 4: Standalone form pages (Observable new, Observable detail, New case form)

- **Observable detail:** Already uses `.sn-form-container` with proportional width; ensure it uses the new variable: `max-width: var(--sn-content-width)` (or keep `min(900px, 92vw)` and align value with the variable).
- **Observable new:** Remove inline `style={{ maxWidth: 640 }}`; use `.sn-form-container` (or a form-specific class) with `max-width: var(--sn-form-width)`.
- **New case form:** Replace fixed `max-width: 960px` with the variable-driven value so it scales with viewport (e.g. `var(--sn-content-width)` or a 960 cap with vw).

### Step 5: List and home pages

- **Case list:** Remove inline `maxWidth: 960`; wrap content in `.sn-content-proportional` or give the existing wrapper `max-width: var(--sn-content-width)` and `margin: 0 auto`.
- **Home:** Same idea; replace inline `maxWidth: 600` with a class that uses `var(--sn-form-width)` or a 600-cap proportional width.

### Step 6: Tables and overflow

- For any `.sn-related-list-panel` or similar that contains a table:
  - Add `overflow-x: auto` (and optionally `-webkit-overflow-scrolling: touch`) so wide tables don’t break the layout on small screens.
- Optionally set `min-width` on the table (e.g. 600px) so columns don’t collapse too much when scrolling.

### Step 7: Optional – narrow viewport (e.g. &lt; 768px)

- **Single-column forms:** Where we use a two-column grid (e.g. `.sn-form-row`, `.sn-form-two-cols`, `.sn-new-record-form`), add a media query so that below 768px the grid becomes one column: `grid-template-columns: 1fr`.
- **Sidebar:** As above, consider hiding or collapsing the case form sidebar so main content gets full width.

---

## 5. Testing Checklist (after each step)

- [ ] Case list: width scales with window; content centered; not overly wide on large screen.
- [ ] Case form (with sidebar): main content has sensible max-width; related list tables scroll horizontally on narrow width.
- [ ] Observable detail: card remains proportional (already done); no regression.
- [ ] Observable new: container uses proportional width; no inline maxWidth override.
- [ ] New case form: two columns scale; on narrow viewport optionally stack to one column.
- [ ] Home: content proportional and centered.
- [ ] No horizontal scroll on the whole page unless it’s from an intentional table scroll.
- [ ] All existing actions (buttons, links, form submit) still work.

---

## 6. File Touch List

| File | Changes |
|------|--------|
| `frontend/src/index.css` | Add variables; add `.sn-content-proportional` (and optional form variant); optional media queries for grids and sidebar; table overflow. |
| `frontend/src/App.css` | If needed, adjust `.app-main` (e.g. max-width or padding). |
| `frontend/src/pages/CaseList.tsx` | Replace inline maxWidth with class. |
| `frontend/src/pages/Home.tsx` | Replace inline maxWidth with class. |
| `frontend/src/pages/ObservableNewPage.tsx` | Remove inline maxWidth; rely on `.sn-form-container` or shared class. |
| `frontend/src/pages/ObservableDetailPage.tsx` | Optional: switch to variable in CSS only. |
| `frontend/src/pages/NewCaseForm.tsx` | Use proportional class on form wrapper if needed. |
| `frontend/src/pages/CaseForm.tsx` | Optional: wrap main content in proportional wrapper; ensure table overflow. |

---

## 7. Order of Work

1. **index.css:** Variables + `.sn-content-proportional` + table overflow.
2. **Case list + Home:** Apply proportional wrapper; remove inline widths.
3. **Observable new:** Use shared form container; remove inline maxWidth.
4. **New case form:** Proportional max-width.
5. **Case form:** Main content max-width + table overflow; optional media query for two-column → one-column and sidebar.
6. **Final pass:** Run through testing checklist; fix any regressions.

This keeps the UI proportional to screen size without breaking layout or behavior.
