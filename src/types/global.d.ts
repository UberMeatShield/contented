// Global declarations
declare var $: any;

// For Monaco editor from window
interface Window {
  monaco: typeof import('monaco-editor');
} 

export type PageResponse<T> = {
  total: number;
  results: T[];
  initialized: boolean;
};