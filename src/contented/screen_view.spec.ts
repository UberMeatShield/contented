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

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';

declare var $;
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
    let screen = new Screen({ id: 'a' });
    expect(screen.url).withContext('It should set the link if possible.').toBeDefined();
  });

  it('Should build out a screen view and be able to render', () => {
    expect(el).withContext('We should have built out a component.').toBeDefined();
    expect($('.screens-cmp').length).withContext('The component should exist').toEqual(1);
  });

  it('Given a content id it will try and render screens', fakeAsync(() => {
    let contentId = 32;
    comp.contentId = contentId;
    fixture.detectChanges();
    expect(comp.loading).toBeTrue();

    let url = ApiDef.contented.contentScreens.replace('{mcID}', contentId.toString());
    let req = httpMock.expectOne(req => req.url == url);
    let sRes = MockData.getScreens();

    expect(sRes.results.length).withContext('We should have screens in the mock data').toBeGreaterThan(0);
    req.flush(sRes);
    tick(1000);

    fixture.detectChanges();
    let expectCount = sRes.results.length;
    expect(comp.loading).toBeFalse(); // It should no longer be loading
    expect(comp.screens.length).withContext('We should have assigned screens').toEqual(expectCount);
    expect($('.screen-img', el).length).withContext('There should be screens rendered').toEqual(expectCount);
    expect($('.screen', el).length).withContext('There should be screens rendered').toEqual(expectCount);
  }));

  it('Can parse out a screen time', () => {
    let s = new Screen({
      id: 'a',
      src: 'Something.ss003.jpg',
    });
    expect(s.parseSecondsFromScreen()).toEqual(3, 'It should parse a second');

    let s2 = new Screen({
      id: 'fake',
      src: 'NoMatch.jpg',
    });
    expect(s2.parseSecondsFromScreen()).toEqual(0, 'No time should not error');
  });
});
