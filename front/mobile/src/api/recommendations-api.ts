import { http } from './http-client';
import { Recommendation, RecommendationSection } from './types/recommendations';

export const recommendationsApi = {
  list: (section: RecommendationSection) =>
    http.get<Recommendation[]>('/recommendations', { query: { section } }),
  dismiss: (workKey: string) =>
    http.post<void>(`/recommendations/${encodeURIComponent(workKey)}/dismiss`),
};
