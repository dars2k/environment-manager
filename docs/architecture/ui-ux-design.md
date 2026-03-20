# UI/UX Design Guide

## Design Philosophy

1. **Dark Theme First** — Modern, eye-friendly dark interface
2. **Information Density** — Display critical data without clutter
3. **Real-time Feedback** — Instant visual updates for all status changes
4. **Role Clarity** — Admin-only actions are hidden, not just disabled, for non-admin users
5. **Action Clarity** — Clear, intuitive controls with visual hierarchy
6. **Responsive Design** — Optimized for desktop and tablet

## Color Palette

```css
/* Primary */
--color-primary:       #3B82F6;   /* Blue — primary actions */
--color-primary-hover: #2563EB;   /* Darker blue — hover */

/* Status */
--color-success: #10B981;   /* Green — healthy */
--color-warning: #F59E0B;   /* Amber — warning / unknown */
--color-danger:  #EF4444;   /* Red — unhealthy / error */
--color-info:    #06B6D4;   /* Cyan — informational */

/* Backgrounds */
--color-bg-primary:   #0F172A;   /* Main background */
--color-bg-secondary: #1E293B;   /* Card backgrounds */
--color-bg-tertiary:  #334155;   /* Elevated surfaces */

/* Text */
--color-text-primary:   #F8FAFC;   /* Primary */
--color-text-secondary: #CBD5E1;   /* Secondary */
--color-text-muted:     #94A3B8;   /* Muted */

/* Borders */
--color-border:       #334155;   /* Default */
--color-border-focus: #3B82F6;   /* Focus */
```

## Typography

```css
--font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
--font-mono: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, monospace;

/* Scale */
--text-xs:   0.75rem;   /* 12px */
--text-sm:   0.875rem;  /* 14px */
--text-base: 1rem;      /* 16px */
--text-lg:   1.125rem;  /* 18px */
--text-xl:   1.25rem;   /* 20px */
--text-2xl:  1.5rem;    /* 24px */
--text-3xl:  1.875rem;  /* 30px */
```

## Page Layouts

### Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  [Logo] Environment Manager         [Role badge] [Username]  │
├──────┬──────────────────────────────────────────────────────┤
│ Nav  │  Environments                     [+ New] (admin)    │
│      │  ┌──────────┬──────────┬──────────┬──────────┐      │
│ ○ D  │  │  Env 1   │  Env 2   │  Env 3   │  Env 4   │      │
│ ○ L  │  │  ● hlthy │  ● hlthy │  ✕ unhlt │  ? unkn  │      │
│ ○ U* │  └──────────┴──────────┴──────────┴──────────┘      │
│      │  ┌──────────┬──────────┐                             │
│      │  │  Env 5   │  Env 6   │                             │
│      │  │  ● hlthy │  ● hlthy │                             │
│      │  └──────────┴──────────┘                             │
└──────┴──────────────────────────────────────────────────────┘
* U (Users) shown only for admin role
```

### Environment Card

Each card displays:
- Status indicator (animated pulse for healthy, static for others)
- Environment name and version badge
- Last health check time and response time
- Quick action buttons: **Restart**, **Upgrade**
- Edit / Delete icons (admin only)

```
┌─────────────────────────────────┐
│  ● production-api    [v2.1.0]   │
│  Last check: 2 min ago          │
│  Response: 145ms                │
│  Last upgrade: 3 days ago       │
├─────────────────────────────────┤
│  [Restart]  [Upgrade]  ✏ 🗑 *  │
└─────────────────────────────────┘
* Edit/Delete shown to admin only
```

### Environment Details

```
┌──────────────────────────────────────────────────────────────┐
│  [← Back]  production-api              [Edit] [Delete] (admin)│
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────────────┬───────────────────────────────┐   │
│  │ Status Overview       │ System Information             │   │
│  │ ● Healthy             │ OS: Ubuntu 22.04 LTS           │   │
│  │ Last check: 2m ago   │ App Version: 2.1.0             │   │
│  │ Response: 145ms      │ IP: 192.168.1.100              │   │
│  └──────────────────────┴───────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Health Check Config                                   │   │
│  │ Endpoint: GET /health   Interval: 30s   Code: 200    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ Recent Logs                                           │   │
│  │ [filterable log timeline]                             │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

### Create / Edit Environment Form

Tabbed layout:

```
Basic Info | Connection | Health Check | Commands | Upgrade Config

┌──────────────────────────────────────────────────────────────┐
│  Name: [________________________]                            │
│  Description: [______________________________________________]│
│  Environment URL: [https://example.com]                      │
│                                                              │
│                              [Cancel]  [Save]                │
└──────────────────────────────────────────────────────────────┘
```

### Users Page (admin only)

```
┌──────────────────────────────────────────────────────────────┐
│  Users                                      [+ Create User]  │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐  │
│  │ Username │  Role    │  Status  │ Last login│  Actions │  │
│  ├──────────┼──────────┼──────────┼──────────┼──────────┤  │
│  │ admin    │ [Admin]  │ ● Active │ 2min ago  │  ✏  🗑  │  │
│  │ alice    │ [User]   │ ● Active │ 1hr ago   │  ✏  🗑  │  │
│  │ bob      │ [Viewer] │ ○ Disabled│ 3d ago   │  ✏  🗑  │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘  │
└──────────────────────────────────────────────────────────────┘
```

## Status Indicators

Animated pulse for live status:

```css
.status-healthy {
  background: #10B981;
  animation: pulse-green 2s infinite;
}

@keyframes pulse-green {
  0%   { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70%  { box-shadow: 0 0 0 10px rgba(16, 185, 129, 0); }
  100% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

.status-unhealthy { background: #EF4444; }
.status-unknown   { background: #F59E0B; }
```

## Real-time Updates

1. WebSocket connection established after login
2. Status changes → `status_update` message → Redux `updateEnvironmentStatus` action → card re-renders
3. Operation progress → `operation_update` message → toast notification
4. Auto-reconnect with exponential backoff on disconnect

## Animations

```css
/* Hover lift for environment cards */
.env-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease;
}

/* Status badge transition */
.status-indicator {
  transition: background-color 0.3s ease;
}
```

## Loading States

| Scenario | Loading pattern |
|----------|----------------|
| Initial page load | Skeleton cards |
| Button action (restart/upgrade) | Spinner inside button |
| Long-running operation | Progress indication via WebSocket |
| Content placeholder | Shimmer effect |

## Error Handling

Errors are surfaced as:

- **Toast notifications** (top-right, auto-dismiss after 5s) — for transient errors (API failures, operation errors)
- **Inline error messages** — for form validation
- **Empty state views** — when lists return no data

## Accessibility

- All interactive elements have `aria-label` attributes
- Role chips use color + text (not color alone) to convey meaning
- Keyboard navigation: `Tab`, `Enter`/`Space`, `Escape` for modals, arrow keys in lists
- All text meets WCAG AA contrast ratios (4.5:1 normal text, 3:1 large text)

## Responsive Breakpoints

```css
/* MUI breakpoints used throughout */
xs:  0px       /* mobile */
sm:  600px     /* tablet */
md:  900px     /* small desktop */
lg:  1200px    /* desktop */
xl:  1536px    /* large desktop */

/* Environment card grid */
xs → 1 column
sm → 2 columns
md → 3 columns
lg → 4 columns
```
