// Mirror of back/api/internal/modules/recommendations/ui/dtos/recommendation-response.go.

export type RecommendationSection = 'for_you' | 'new' | 'discover';

export interface Recommendation {
  workKey: string;
  title: string;
  authors: string[];
  /** First publication year; 0 when unknown. */
  year: number;
  coverUrl: string;
  subjects: string[];
  /** The user's book closest to this one. */
  reason: { workKey: string; title: string } | null;
}
