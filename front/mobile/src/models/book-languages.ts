/** Languages offered in the library search, keyed by ISO code, with the name libgen uses. */
export const bookLanguages = {
  es: 'Spanish',
  en: 'English',
  fr: 'French',
  de: 'German',
  it: 'Italian',
  pt: 'Portuguese',
  ca: 'Catalan',
  ru: 'Russian',
  ja: 'Japanese',
  zh: 'Chinese',
  nl: 'Dutch',
  pl: 'Polish',
  tr: 'Turkish',
} as const;

export type BookLanguage = keyof typeof bookLanguages;

export const bookLanguageCodes = Object.keys(bookLanguages) as BookLanguage[];

export function isBookLanguage(code: string): code is BookLanguage {
  return code in bookLanguages;
}

/** Codes of the languages of a book as libgen writes them ("Spanish; English"). */
export function bookLanguagesOf(libgenLanguage: string): (BookLanguage | string)[] {
  return libgenLanguage
    .split(/[,;/]/)
    .map((l) => l.trim())
    .filter(Boolean)
    .map(
      (l) => bookLanguageCodes.find((c) => bookLanguages[c].toLowerCase() === l.toLowerCase()) ?? l
    );
}
