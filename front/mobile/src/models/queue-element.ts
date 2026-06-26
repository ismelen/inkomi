import { Source } from './source';
import { Destination } from './transaction-request';

export interface QueueElement {
  timestamp: number;
  filename: string;
  id: string;
  title: string;
  destination: Destination;
  progress: number;
  error?: string;
  sources: Source[];
}
