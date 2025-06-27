# UI/UX Design Guide

## Design Philosophy

The Application Environment Manager UI follows these core principles:

1. **Dark Theme First**: Modern, eye-friendly dark interface
2. **Information Density**: Display critical data without clutter
3. **Real-time Feedback**: Instant visual updates for all status changes
4. **Action Clarity**: Clear, intuitive controls for critical operations
5. **Responsive Design**: Works seamlessly on desktop and tablet

## Visual Design System

### Color Palette

```css
/* Primary Colors */
--color-primary: #3B82F6;        /* Blue - Primary actions */
--color-primary-hover: #2563EB;  /* Darker blue - Hover state */

/* Status Colors */
--color-success: #10B981;        /* Green - Healthy status */
--color-warning: #F59E0B;        /* Amber - Warning/Unknown */
--color-danger: #EF4444;         /* Red - Unhealthy/Error */
--color-info: #06B6D4;          /* Cyan - Information */

/* Background Colors */
--color-bg-primary: #0F172A;     /* Main background */
--color-bg-secondary: #1E293B;   /* Card backgrounds */
--color-bg-tertiary: #334155;    /* Elevated surfaces */

/* Text Colors */
--color-text-primary: #F8FAFC;   /* Primary text */
--color-text-secondary: #CBD5E1; /* Secondary text */
--color-text-muted: #94A3B8;     /* Muted text */

/* Border Colors */
--color-border: #334155;         /* Default borders */
--color-border-focus: #3B82F6;   /* Focus borders */
```

### Typography

```css
/* Font Stack */
--font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, 
             "Helvetica Neue", Arial, sans-serif;
--font-mono: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", 
             Consolas, monospace;

/* Font Sizes */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
```

## Component Library

### 1. Environment Card

```jsx
<EnvironmentCard>
  <CardHeader>
    <StatusIndicator status="healthy" />
    <EnvironmentName>production-api</EnvironmentName>
    <VersionBadge>v2.1.0</VersionBadge>
  </CardHeader>
  
  <CardBody>
    <MetricRow>
      <MetricLabel>Last Check</MetricLabel>
      <MetricValue>2 min ago</MetricValue>
    </MetricRow>
    <MetricRow>
      <MetricLabel>Response Time</MetricLabel>
      <MetricValue>145ms</MetricValue>
    </MetricRow>
    <MetricRow>
      <MetricLabel>Last Upgrade</MetricLabel>
      <MetricValue>3 days ago</MetricValue>
    </MetricRow>
  </CardBody>
  
  <CardActions>
    <ActionButton variant="primary" icon="restart">
      Restart
    </ActionButton>
    <ActionButton variant="secondary" icon="shutdown">
      Shutdown
    </ActionButton>
    <ActionButton variant="info" icon="upgrade">
      Upgrade
    </ActionButton>
  </CardActions>
</EnvironmentCard>
```

### 2. Status Indicator

Real-time animated status indicator with pulse effect:

```css
/* Healthy Status */
.status-indicator.healthy {
  background: #10B981;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 1);
  animation: pulse-green 2s infinite;
}

@keyframes pulse-green {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}
```

### 3. Action Buttons

Three button variants with clear visual hierarchy:

```jsx
<Button variant="primary">Primary Action</Button>
<Button variant="secondary">Secondary Action</Button>
<Button variant="danger">Dangerous Action</Button>
```

## Page Layouts

### Dashboard Layout

```
┌─────────────────────────────────────────────────────────────┐
│  [Logo] Application Environment Manager     [User] [Settings]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Environments (12 total)          [+ Add Environment]       │
│  ┌──────────┬──────────┬──────────┬──────────┐           │
│  │  Env 1   │  Env 2   │  Env 3   │  Env 4   │           │
│  │  Card    │  Card    │  Card    │  Card    │           │
│  └──────────┴──────────┴──────────┴──────────┘           │
│  ┌──────────┬──────────┬──────────┬──────────┐           │
│  │  Env 5   │  Env 6   │  Env 7   │  Env 8   │           │
│  │  Card    │  Card    │  Card    │  Card    │           │
│  └──────────┴──────────┴──────────┴──────────┘           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Environment Details Layout

```
┌─────────────────────────────────────────────────────────────┐
│  [← Back] production-api                    [Edit] [Delete] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────┬───────────────────────────────┐  │
│  │ Status Overview     │ System Information             │  │
│  │                     │                                │  │
│  │ ● Healthy          │ OS: Ubuntu 22.04 LTS           │  │
│  │ Last check: 2m ago │ App Version: 2.1.0             │  │
│  │ Response: 145ms    │ IP: 192.168.1.100              │  │
│  │ Uptime: 99.95%     │ Domain: api.example.com        │  │
│  └─────────────────────┴───────────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ Health Check Configuration                           │  │
│  │                                                      │  │
│  │ Endpoint: GET /health                               │  │
│  │ Interval: Every 30 seconds                          │  │
│  │ Validation: Status Code 200                         │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ Recent Activity                                      │  │
│  │                                                      │  │
│  │ [Timeline of recent actions and status changes]      │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Create/Edit Environment Form

```
┌─────────────────────────────────────────────────────────────┐
│  [← Cancel] Create New Environment                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Basic Information                                          │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ Name: [____________________]                        │  │
│  │ Description: [_____________________________________]│  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  Connection Details                                         │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ Host/IP: [____________________] Port: [22]         │  │
│  │ Domain (optional): [____________________]          │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  SSH Credentials                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ Username: [____________________]                    │  │
│  │ Auth Type: (•) SSH Key  ( ) Password               │  │
│  │ [Upload SSH Key] or paste below:                   │  │
│  │ [_________________________________________________]│  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  Health Check Configuration                                 │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ [✓] Enable Health Checks                            │  │
│  │ Endpoint: [____________________]                    │  │
│  │ Method: [GET ▼]  Interval: [30 ▼] seconds         │  │
│  │ Validation: (•) Status Code [200]                  │  │
│  │            ( ) JSON Regex [____________________]   │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  [Cancel]                              [Test] [Save]        │
└─────────────────────────────────────────────────────────────┘
```

## Real-time Updates Strategy

### WebSocket Connection

1. **Connection Management**
   ```javascript
   // Auto-reconnect with exponential backoff
   const reconnectDelays = [1000, 2000, 4000, 8000, 16000];
   ```

2. **Update Types**
   - Status changes: Immediate card animation
   - Metrics updates: Smooth transitions
   - Operation progress: Real-time progress bars

3. **Visual Feedback**
   - Pulse animation for incoming updates
   - Fade transitions for status changes
   - Loading skeletons during reconnection

### Optimistic UI Updates

```javascript
// Immediate UI update
updateEnvironmentUI(envId, { status: 'restarting' });

// Rollback on failure
api.restart(envId).catch(() => {
  rollbackEnvironmentUI(envId);
  showError('Restart failed');
});
```

## Responsive Design

### Breakpoints

```css
/* Mobile: 320px - 768px */
@media (max-width: 768px) {
  /* Single column layout */
}

/* Tablet: 768px - 1024px */
@media (min-width: 768px) and (max-width: 1024px) {
  /* 2-column grid */
}

/* Desktop: 1024px - 1440px */
@media (min-width: 1024px) and (max-width: 1440px) {
  /* 3-column grid */
}

/* Large Desktop: 1440px+ */
@media (min-width: 1440px) {
  /* 4-column grid */
}
```

## Accessibility

### ARIA Labels

```jsx
<button 
  aria-label="Restart production-api environment"
  aria-pressed={isRestarting}
  disabled={isRestarting}
>
  {isRestarting ? 'Restarting...' : 'Restart'}
</button>
```

### Keyboard Navigation

- `Tab`: Navigate between interactive elements
- `Enter/Space`: Activate buttons
- `Escape`: Close modals/dropdowns
- `Arrow keys`: Navigate within lists

### Color Contrast

All text meets WCAG AA standards:
- Normal text: 4.5:1 contrast ratio
- Large text: 3:1 contrast ratio

## Animation Guidelines

### Micro-interactions

```css
/* Button hover */
.button {
  transition: all 0.2s ease;
}

/* Card hover */
.env-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* Status change */
.status-indicator {
  transition: background-color 0.3s ease;
}
```

### Loading States

1. **Skeleton Screens**: For initial page loads
2. **Spinner**: For button actions
3. **Progress Bar**: For long operations
4. **Shimmer Effect**: For content placeholders

## Error Handling

### Error Messages

```jsx
<ErrorMessage type="warning">
  <Icon name="warning" />
  <Title>Connection Failed</Title>
  <Description>
    Unable to connect to production-api. 
    Check your network settings.
  </Description>
  <Actions>
    <Button onClick={retry}>Retry</Button>
    <Button onClick={viewDetails}>View Details</Button>
  </Actions>
</ErrorMessage>
```

### Toast Notifications

Position: Top-right corner
Duration: 5 seconds (auto-dismiss)
Types: Success, Warning, Error, Info

## Performance Considerations

1. **Virtual Scrolling**: For large environment lists
2. **Lazy Loading**: For detail views and charts
3. **Debounced Search**: 300ms delay
4. **Memoized Components**: For expensive renders
5. **Code Splitting**: Route-based chunks
