import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync, flush } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { Location } from '@angular/common';
import { DebugElement, NgZone } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { provideRouter, Router } from '@angular/router';

import { SearchCmp } from '../contented/search.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule, WaitForMonacoLoad } from '../contented/contented_module';
import { Container } from '../contented/container';
import { ApiDef } from '../contented/api_def';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';
import { RouterTestingHarness } from '@angular/router/testing';
import { Content, ContentSchema, Tag, VideoCodecInfo, VideoCodecInfoSchema, VideoFormatSchema, VideoStreamSchema } from './content';

describe('TestingSearchCmp', () => {
  let fixture: ComponentFixture<SearchCmp>;
  let service: ContentedService;
  let comp: SearchCmp;
  let router: Router;

  let httpMock: HttpTestingController;
  let loc: Location;
  let harness: RouterTestingHarness;

  beforeEach(waitForAsync(async () => {
    TestBed.configureTestingModule({
      imports: [FormsModule, ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [provideRouter([{ path: 'ui/search/', component: SearchCmp }]), ContentedService],
      teardown: { destroyAfterEach: true },
    }).compileComponents();

    harness = await RouterTestingHarness.create();

    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
    loc = TestBed.inject(Location);
    router = TestBed.inject(Router);
    router.initialNavigation();

    spyOn(console, 'error');
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a search component', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/search/?searchText=Cthulhu', SearchCmp);
    harness.detectChanges();
    let de: DebugElement = harness.fixture.debugElement.query(By.css('.search-cmp'));
    let el: HTMLElement = de.nativeElement;

    expect(comp).withContext('We should have the Contented comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  }));

  it("Should be able to create content for all search results without error", () => {
    let sr = MockData.getSearch();

    const first = sr.results[0];
    expect(first.content_type).toBe('video/mp4');
    const meta = JSON.parse(first.meta);
    const probe = VideoCodecInfoSchema.parse(meta)
    //expect(stream.success).toBe(true);

    for (const r of sr.results) {
      const content = new Content(r);
    }
  });

  it('It can setup all eventing without exploding', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/search/?searchText=Cthulhu', SearchCmp);
    harness.detectChanges();
    expect(comp.searchText).withContext('It should default via route params').toBe('Cthulhu');

    comp.search(comp.searchText, 0, 50, null);
    let sr = MockData.getSearch();
    expect(sr.results.length).withContext('We need some search results.').toBeGreaterThan(0);

    const reqs = httpMock.match(req => req.url === ApiDef.contented.searchContents);
    expect(reqs.length).toBe(1);
    reqs.forEach(req => req.flush(sr));
    harness.detectChanges();

    expect(comp.loading).toBeFalse();
    harness.detectChanges();
    expect(comp.content.length).withContext('It should have results').toEqual(sr.results.length);
    expect(comp.loading).toBeFalse();
  }));
});
