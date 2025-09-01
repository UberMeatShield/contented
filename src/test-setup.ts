import '@analogjs/vitest-angular/setup-zone';

import { BrowserDynamicTestingModule, platformBrowserDynamicTesting } from '@angular/platform-browser-dynamic/testing';
import { getTestBed } from '@angular/core/testing';

// analogjs stops an annoying warning about angular material styles in tests
// https://github.com/analogjs/analog/issues/1673
Object.defineProperty(window, 'getComputedStyle', {
  value: () => {
    return {
      display: 'none',
    };
  },
});

(window as any).VITEST = true;

getTestBed().initTestEnvironment(BrowserDynamicTestingModule, platformBrowserDynamicTesting());
