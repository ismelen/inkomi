import { create } from 'zustand';
import { errorKey } from '../api/api-error';
import { recommendationsApi } from '../api/recommendations-api';
import { Recommendation, RecommendationSection } from '../api/types/recommendations';

type Status = 'idle' | 'loading' | 'done' | 'error';

interface RecommendationsState {
  section: RecommendationSection;
  bySection: Partial<Record<RecommendationSection, Recommendation[]>>;
  status: Status;
  /** i18n key of the last error. */
  error?: string;
  setSection(section: RecommendationSection): void;
  load(force?: boolean): Promise<void>;
  dismiss(workKey: string): Promise<string | undefined>;
}

export const useRecommendations = create<RecommendationsState>()((set, get) => ({
  section: 'for_you',
  bySection: {},
  status: 'idle',

  setSection: (section) => {
    set({ section });
    void get().load();
  },

  load: async (force = false) => {
    const section = get().section;
    if (!force && get().bySection[section]) {
      set({ status: 'done', error: undefined });
      return;
    }
    set({ status: 'loading', error: undefined });
    try {
      const recs = await recommendationsApi.list(section);
      set((s) => ({ bySection: { ...s.bySection, [section]: recs }, status: 'done' }));
    } catch (err) {
      set({ status: 'error', error: errorKey(err) });
    }
  },

  dismiss: async (workKey) => {
    const previous = get().bySection;
    const without: RecommendationsState['bySection'] = {};
    (Object.keys(previous) as RecommendationSection[]).forEach((section) => {
      without[section] = previous[section]?.filter((rec) => rec.workKey !== workKey);
    });
    set({ bySection: without });
    try {
      await recommendationsApi.dismiss(workKey);
      return undefined;
    } catch (err) {
      set({ bySection: previous });
      return errorKey(err);
    }
  },
}));
