import { TransactionRequest } from './transaction-request';

export interface Upload {
  request: TransactionRequest;
  timestamp: number;
  id: string;
  error?: Error;
  libgenMode: boolean;
}
