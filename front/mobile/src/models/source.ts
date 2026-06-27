export interface Source {
  name: string;
  path: string;
  size?: number;
  mime?: string;
  children?: Source[];
}
