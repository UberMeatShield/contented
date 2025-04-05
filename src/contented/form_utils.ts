/**
 * Utility functions for working with Angular Reactive Forms in strict TypeScript mode
 */

import { FormControl, FormGroup, Validators, ValidatorFn } from '@angular/forms';

/**
 * Creates a non-nullable FormControl that requires a default value
 * This is used when you need a FormControl<T> rather than FormControl<T | null>
 */
export function createRequiredControl<T>(value: T, validators: ValidatorFn[] = []): FormControl<T> {
  return new FormControl<T>(value, { nonNullable: true, validators }) as FormControl<T>;
}

/**
 * Creates a nullable FormControl that can be null or undefined
 * This is useful when the field is optional
 */
export function createOptionalControl<T>(value: T | null = null, validators: ValidatorFn[] = []): FormControl<T | null> {
  return new FormControl<T | null>(value, validators);
}

/**
 * Creates a string FormControl that is guaranteed to be a string
 * Useful for text inputs where null is not acceptable
 */
export function createStringControl(value: string = '', validators: ValidatorFn[] = []): FormControl<string> {
  return createRequiredControl<string>(value, validators);
}

/**
 * Creates a numeric FormControl that is guaranteed to be a number
 * Useful for numeric inputs where null is not acceptable
 */
export function createNumberControl(value: number = 0, validators: ValidatorFn[] = []): FormControl<number> {
  return createRequiredControl<number>(value, validators);
}

/**
 * Creates a boolean FormControl that is guaranteed to be a boolean
 * Useful for checkboxes and toggles
 */
export function createBooleanControl(value: boolean = false, validators: ValidatorFn[] = []): FormControl<boolean> {
  return createRequiredControl<boolean>(value, validators);
} 