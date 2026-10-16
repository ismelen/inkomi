// Open Color palette of the "Mobile" frame in docs/ideas.excalidraw.
export const colors = {
  text: '#1e1e1e',
  textSecondary: '#868e96',
  border: '#ced4da',
  surface: '#e9ecef',
  background: '#ffffff',
  primary: '#1e1e1e',
  onPrimary: '#ffffff',
  success: '#2f9e44',
  info: '#1971c2',
  warning: '#f08c00',
  danger: '#e03131',
  scrim: 'rgba(30, 30, 30, 0.45)',
  ripple: 'rgba(30, 30, 30, 0.08)',
} as const;

export type ColorToken = keyof typeof colors;
