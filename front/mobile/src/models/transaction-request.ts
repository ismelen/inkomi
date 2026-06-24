import { Source } from './source';

export interface TransactionRequest {
  sources: Source[];
  title?: string;
  author?: string;
  destination: Destination;
  merge?: boolean;
  deleteOrigin: boolean;
}

export type Destination = 'local' | 'cloud';
