export interface Source {
  name: string;
  path: string;
  size?: number;
  children?: Source[];
}
