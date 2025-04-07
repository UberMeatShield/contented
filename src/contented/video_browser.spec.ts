import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { Location } from '@angular/common';
import { DebugElement, NgZone } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingHarness } from '@angular/router/testing';
import { provideRouter, Router } from '@angular/router';

import { VideoBrowserCmp } from '../contented/video_browser.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { ApiDef } from '../contented/api_def';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';
import { ContainerSchema } from './container';

describe('TestingVideoBrowserCmp', () => {
  let fixture: ComponentFixture<VideoBrowserCmp>;
  let service: ContentedService;
  let comp: VideoBrowserCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;
  let loc: Location;
  let harness: RouterTestingHarness;

  beforeEach(waitForAsync(async () => {
    TestBed.configureTestingModule({
      imports: [FormsModule, ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [ContentedService, provideRouter([{ path: 'ui/video/', component: VideoBrowserCmp }])],
      teardown: { destroyAfterEach: true },
    }).compileComponents();

    harness = await RouterTestingHarness.create();
    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
    loc = TestBed.inject(Location);
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it("Should load containers", () => {
    const containerResult = MockData.getContainers();

    containerResult.results.forEach(c => {
      const container = ContainerSchema.safeParse(c);
      expect(container.success).withContext(container?.error?.message).toBe(true);
    });
  });

  it('Should create a video component', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/video/', VideoBrowserCmp);
    de = harness.fixture.debugElement.query(By.css('.video-browser-cmp'));
    el = de.nativeElement;
    expect(comp).withContext('We should have the VideoBrowser comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();

    httpMock.expectOne(ApiDef.contented.containers, 'Empty').flush({total: 0, results: [] });
  }));

  it('It can setup all eventing without exploding', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/video/?searchText=Cthulhu', VideoBrowserCmp);
    harness.detectChanges();
    expect(comp.searchText).toEqual('Cthulhu');
    const values = comp.getValues();
    expect(values['searchType']).toBe('text');
    harness.detectChanges();

    MockData.handleContainerLoad(httpMock);
    comp.search('Cthulhu', 0, 50, '1');

    let req = httpMock.expectOne(req => req.url === ApiDef.contented.searchContents, 'Failed to find search');
    let searchResults = MockData.getVideos();

    expect(searchResults.results.length).withContext('We need some search results.').toBeGreaterThan(0);
    req.flush(searchResults);
    harness.detectChanges();
    expect($('.video-view-card').length).toEqual(searchResults.results.length);

    for (const content of searchResults.results) {
      let screenUrl = ApiDef.contented.contentScreens.replace('{mcID}', content.id.toString());
      let screenReq = httpMock.expectOne(req => req.url.includes(screenUrl));
      screenReq.flush(MockData.getScreens());
    }
    harness.detectChanges();
  }));

  it('Will load up screens if they are not provided', waitForAsync(async () => {
    let vRes = MockData.getVideos();
    _.each(vRes.results, v => {
      v.screens = null;
    });
    comp = await harness.navigateByUrl('/ui/video/?searchText=Cthulhu', VideoBrowserCmp);
    harness.detectChanges();
    MockData.handleContainerLoad(httpMock);

    comp.search('Cthulhu', 0, 50, 'A');
    harness.detectChanges();
    let req = httpMock.expectOne(req => req.url === ApiDef.contented.searchContents);
    req.flush(vRes);
    harness.detectChanges();

    _.each(vRes.results, content => {
      let screenUrl = ApiDef.contented.contentScreens.replace('{mcID}', content.id.toString());
      let screenReq = httpMock.expectOne(req => req.url.includes(screenUrl));
      screenReq.flush(MockData.getScreens());
    });
    harness.detectChanges();
    expect($('.video-details').length).withContext('Should show 4 details sections').toEqual(4);
  }));
});
