import { Source } from './source';
import { Destination } from './transaction-request';

export interface QueueElement {
  id: string;
  title: string;
  destination: Destination;
  progress: number;
  error?: string;
  sources: Source[];
}
