import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { Screen } from '../contented/screen';
import { ScreensCmp } from '../contented/screens.cmp';
import { Container } from '../contented/container';
import { ApiDef } from '../contented/api_def';
import { GlobalNavEvents } from '../contented/nav_events';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';
import { describe } from 'vitest';

describe('TestingScreensCmp', () => {
  let fixture: ComponentFixture<ScreensCmp>;
  let service: ContentedService;
  let comp: ScreensCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule.withRoutes([{ path: 'screens/:screenId', component: ScreensCmp }]),
        FormsModule,
        ContentedModule,
        HttpClientTestingModule,
      ],
      providers: [ContentedService],
    }).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(ScreensCmp);
    httpMock = TestBed.get(HttpTestingController);
    comp = fixture.componentInstance;
    de = fixture.debugElement.query(By.css('.screens-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a screens view component', () => {
    const screens = MockData.getScreens();
    for (const s of screens.results) {
      try {
        const screen = new Screen(s);
        expect(screen.url).withContext('It should set the link if possible.').toBeDefined();
        if (screen.idx > 0) {
          expect(screen.timeSeconds).withContext('It should set the time if possible.').toBeGreaterThan(0);
        }
      } catch (e) {
        console.error(`${e}`);
      }
    }
  });

  it('Should build out a screen view and be able to render', () => {
    expect(el).withContext('We should have built out a component.').toBeDefined();
    expect($('.screens-cmp').length).withContext('The component should exist').toEqual(1);
  });

  it('Given a content id it will try and render screens', fakeAsync(() => {
    let contentId = 32;
    comp.contentId = contentId;
    fixture.detectChanges();
    expect(comp.loading).toBe(true);

    let url = ApiDef.contented.contentScreens.replace('{mcID}', contentId.toString());
    let req = httpMock.expectOne(req => req.url == url);
    let sRes = MockData.getScreens();

    expect(sRes.results.length).withContext('We should have screens in the mock data').toBeGreaterThan(0);
    req.flush(sRes);
    tick(1000);

    fixture.detectChanges();
    let expectCount = sRes.results.length;
    expect(comp.loading).toBe(false); // It should no longer be loading
    expect(comp.screens.length).withContext('We should have assigned screens').toEqual(expectCount);
    expect($('.screen-img', el).length).withContext('There should be screens rendered').toEqual(expectCount);
    expect($('.screen', el).length).withContext('There should be screens rendered').toEqual(expectCount);
  }));

  it('Can parse out a screen time', () => {
    const sample = {
      id: 10,
      src: 'Something.ss003.jpg',
      content_id: 1,
      content_container_id: 1,
    };
    const s = new Screen(sample);
    expect(s.parseSecondsFromScreen()).toEqual(3, 'It should parse a second');

    let s2 = new Screen({
      id: 20,
      src: 'NoMatch.jpg',
      content_id: 1,
    });
    expect(s2.parseSecondsFromScreen()).toEqual(0, 'No time should not error');
  });
});
