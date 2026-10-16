export const spacing = {
  xxs: 2,
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
} as const;

export const radii = {
  sm: 6,
  md: 10,
  lg: 16,
  pill: 999,
} as const;

export const fonts = {
  regular: 'regular',
  medium: 'medium',
  semibold: 'semibold',
  bold: 'bold',
  black: 'black',
} as const;

export const typography = {
  section: { fontSize: 11, fontFamily: fonts.semibold, letterSpacing: 0.6 },
  small: { fontSize: 12, fontFamily: fonts.regular },
  body: { fontSize: 14, fontFamily: fonts.regular },
  label: { fontSize: 14, fontFamily: fonts.medium },
  title: { fontSize: 16, fontFamily: fonts.semibold },
  h1: { fontSize: 20, fontFamily: fonts.bold },
  brand: { fontSize: 20, fontFamily: fonts.black, letterSpacing: 2 },
} as const;

export type TypographyVariant = keyof typeof typography;
