import { Pipe, PipeTransform } from '@angular/core';

// https://gist.github.com/thomseddon/3511330
@Pipe({
  name: 'byteFormatter'
})
export class ByteFormatterPipe implements PipeTransform {

  transform(bytes: string | number, precision: number): any {

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
