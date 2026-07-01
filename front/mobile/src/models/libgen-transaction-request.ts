import { LibgenBook } from './book';
import { TransactionRequest } from './transaction-request';

export interface LibgenTransactionRequest extends TransactionRequest {
  books: LibgenBook[];
}
