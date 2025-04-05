/**
 * Type-safe utility functions for Angular 19 and strict TypeScript
 */

import { Content } from './content';
import { Container } from './container';

/**
 * Safely handles null/undefined values for Content objects
 * @param content The content object or undefined/null
 * @returns A valid Content object or undefined
 */
export function safeContent(content: Content | null | undefined): Content | undefined {
  return content === null ? undefined : content;
}

/**
 * Safely handles null/undefined values for Container objects
 * @param container The container object or undefined/null
 * @returns A valid Container object or undefined
 */
export function safeContainer(container: Container | null | undefined): Container | undefined {
  return container === null ? undefined : container;
}

/**
 * Provides a default value for null/undefined strings
 * @param value The string value to check
 * @param defaultValue The default value to use if value is null/undefined
 * @returns A valid string
 */
export function safeString(value: string | null | undefined, defaultValue: string = ''): string {
  return value === null || value === undefined ? defaultValue : value;
}

/**
 * Type helper for initializing class properties in a type-safe way
 * This helps with strict initialization requirements in Angular components
 */
export interface DefaultValues<T> {
  [key: string]: any;
}

/**
 * Initialize properties with default values to comply with TypeScript's strict property initialization
 * 
 * Example usage:
 * ```
 * @Component({...})
 * export class MyComponent implements OnInit {
 *   @Input() content!: Content;
 *   @Input() width: number;
 *   @Input() height: number;
 *   public loading: boolean;
 *   
 *   constructor() {
 *     initializeDefaults(this, {
 *       width: 0,
 *       height: 0,
 *       loading: false
 *     });
 *   }
 * }
 * ```
 */
export function initializeDefaults<T>(instance: T, defaults: DefaultValues<T>): void {
  Object.keys(defaults).forEach(key => {
    (instance as any)[key] = defaults[key];
  });
} 