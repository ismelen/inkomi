import { router } from 'expo-router';
import { SearchField } from '../api/types/books';
import { useSearch } from '../stores/use-search';

/** Opens the library search already searching text (a title or an author). */
export function openLibrarySearch(text: string, field: SearchField = 'title') {
  const search = useSearch.getState();
  search.setQuery(text);
  useSearch.setState({ field });
  void search.search();
  router.push('/search');
}
