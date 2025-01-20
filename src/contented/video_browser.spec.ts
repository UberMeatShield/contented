import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { Location } from '@angular/common';
import { DebugElement, NgZone } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { VideoBrowserCmp } from '../contented/video_browser.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { ApiDef } from '../contented/api_def';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';

declare var $;
describe('TestingVideoBrowserCmp', () => {
  let fixture: ComponentFixture<VideoBrowserCmp>;
  let service: ContentedService;
  let comp: VideoBrowserCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;
  let loc: Location;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule.withRoutes([{ path: 'ui/video', component: VideoBrowserCmp }]),
        FormsModule,
        ContentedModule,
        HttpClientTestingModule,
        NoopAnimationsModule,
      ],
      providers: [ContentedService],
      teardown: { destroyAfterEach: false },
    }).compileComponents();

    service = TestBed.inject(ContentedService);
    fixture = TestBed.createComponent(VideoBrowserCmp);
    httpMock = TestBed.inject(HttpTestingController);
    loc = TestBed.inject(Location);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.video-browser-cmp'));
    el = de.nativeElement;
    router = TestBed.inject(Router);

    const ngZone = TestBed.inject(NgZone);
    ngZone.run(() => {
      router.initialNavigation();
    });

    // Annoying hack
    comp.tags = [
      {
        id: 'a',
        tag_type: '',
        isProblem: function (): boolean {
          return false;
        },
      },
    ];
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the VideoBrowser comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('It can setup all eventing without exploding', fakeAsync(() => {
    let st = 'Cthulhu';
    router.navigate(['/ui/video/'], { queryParams: { searchText: st } });
    tick(1000);
    fixture.detectChanges();
    tick(1000);
    expect(comp.searchText).toEqual(st);
    const values = comp.getValues();
    expect(values['searchType']).toBe('text');
    fixture.detectChanges();
    tick(1000);

    MockData.handleContainerLoad(httpMock);
    let req = httpMock.expectOne(req => req.url === ApiDef.contented.searchContents, 'Failed to find search');
    let searchResults = MockData.getVideos();

    expect(searchResults.results.length).withContext('We need some search results.').toBeGreaterThan(0);
    req.flush(searchResults);
    fixture.detectChanges();
    expect($('.video-view-card').length).toEqual(searchResults.results.length);
    tick(1000);
    tick(1000);

    for (const content of searchResults.results) {
      let screenUrl = ApiDef.contented.contentScreens.replace('{mcID}', content.id);
      let screenReq = httpMock.expectOne(req => req.url.includes(screenUrl));
      screenReq.flush(MockData.getScreens());
    }

    fixture.detectChanges();
    tick(1000);
    tick(10000);
  }));

  it('Will load up screens if they are not provided', fakeAsync(() => {
    let vRes = MockData.getVideos();
    _.each(vRes.results, v => {
      v.screens = null;
    });
    fixture.detectChanges();
    MockData.handleContainerLoad(httpMock);
    tick(1000);

    let req = httpMock.expectOne(req => {
      console.log('What is the url', req.url);
      return req.url === ApiDef.contented.searchContents;
    }, 'Failed to find search');
    req.flush(vRes);
    fixture.detectChanges();
    tick(1000);

    _.each(vRes.results, content => {
      let screenUrl = ApiDef.contented.contentScreens.replace('{mcID}', content.id);
      let screenReq = httpMock.expectOne(req => req.url.includes(screenUrl));
      screenReq.flush(MockData.getScreens());
    });
    fixture.detectChanges();
    tick(1000);
    fixture.detectChanges();
    expect($('.video-details').length).withContext('Should show 4 details sections').toEqual(4);
    tick(10000);
    tick(10000);
  }));
});
