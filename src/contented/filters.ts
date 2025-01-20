import { Pipe, PipeTransform } from '@angular/core';

// https://gist.github.com/thomseddon/3511330
@Pipe({
  name: 'byteFormatter',
})
export class ByteFormatterPipe implements PipeTransform {
  transform(bytes: string | number, precision: number): any {
    if (bytes === null || bytes === undefined) {
      return '-';
    }

    if (isNaN(parseFloat(bytes.toString())) || !isFinite(Number(bytes))) {
      return '-';
    }

    if (typeof precision === 'undefined') {
      precision = 1;
    }

    const units = ['bytes', 'kB', 'MB', 'GB', 'TB', 'PB'];
    const actualValue = Math.floor(Math.log(Number(bytes)) / Math.log(1024));

    return (Number(bytes) / Math.pow(1024, Math.floor(actualValue))).toFixed(precision) + ' ' + units[actualValue];
  }
}

// From Brave AI results
@Pipe({ name: 'durationFormat' })
export class DurationFormatPipe implements PipeTransform {
  transform(value: number, inputType: 'ms' | 's', format: 'hhmmss' | 'ddhhmmss' | 'ddhhmmssLong' = 'hhmmss'): string {
    if (inputType === 'ms') {
      value = value / 1000;
    }
    const hours = Math.floor(value / 3600);
    const minutes = Math.floor((value % 3600) / 60);
    const seconds = value % 60;

    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }
}
