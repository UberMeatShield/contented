import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { ContentViewCmp } from '../contented/content_view.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { ApiDef } from '../contented/api_def';

import * as _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';

const donutMock = {
  active: false,
  container_id: 22,
  content_type: 'video/mp4',
  corrupt: false,
  created: '0001-01-01T00:00:00Z',
  description: '',
  encoding: '',
  id: 42,
  idx: 1,
  preview: '/container_previews/donut.mp4.webp',
  size: 18401008,
  src: 'donut.mp4',
  updated: '0001-01-01T00:00:00Z',
};

describe('TestingContentViewCmp', () => {
  let fixture: ComponentFixture<ContentViewCmp>;
  let service: ContentedService;
  let comp: ContentViewCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule.withRoutes([{ path: 'ui/view/:id', component: ContentViewCmp }]),
        FormsModule,
        ContentedModule,
        HttpClientTestingModule,
      ],
      providers: [ContentedService],
    }).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(ContentViewCmp);
    httpMock = TestBed.get(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.content-view-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).toBeDefined('We should have the Contented comp');
    expect(el).toBeDefined('We should have a top level element');
  });

  it('Fully handles routing arguments', fakeAsync(() => {
    // Should just setup other ajax calls
    fixture.detectChanges();

    let id = 'FakeUUID';
    let url = `/ui/view/${id}`;
    router.navigate([url]);
    tick(100);
    expect(router.url).toBe(url);
    fixture.detectChanges();

    // TODO: Make a test that actually works with the damn activated route params
    // The route subscription doesn't actually seem to change or happen in tests.
    tick(1000);
    // expect(comp.contentID).toEqual(id);
  }));

  it('Can load up a content ID and will render the correct elements', () => {
    fixture.detectChanges();
    expect($('.content-view-fullscreen').length).toEqual(0, "It shouldn't be visible");
    expect($('.loading').length).toEqual(0, 'Nothing should be loading');

    let fakeID = donutMock.id;
    comp.loadContent(fakeID);
    expect(comp.loading).toBeTrue();
    fixture.detectChanges();
    expect($('.loading').length).toEqual(1, 'Loading UI should be present');

    let url = ApiDef.contented.content.replace('{id}', fakeID.toString());
    let req = httpMock.expectOne(url);
    req.flush(donutMock);
    fixture.detectChanges();
    expect($('.content-view-fullscreen').length).toEqual(1, 'It should now be visible');

    let screenUrl = ApiDef.contented.contentScreens.replace('{mcID}', fakeID.toString());
    let screenReq = httpMock.expectOne(screenUrl);
    let screens = MockData.getScreens();
    screenReq.flush(screens);
    fixture.detectChanges();

    let count = screens.total;
    expect($('.screen-img').length).withContext('We should have screens visible').toEqual(count);
    expect($('.error').length).withContext('No errors should be present').toEqual(0);
    expect($('.loading').length).withContext('Nothing is loading anymore').toEqual(0);
  });
});
