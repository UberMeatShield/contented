import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { Subscription } from 'rxjs';

import { RouterTestingModule } from '@angular/router/testing';
import { ContainerNavCmp } from '../contented/container_nav.cmp';
import { Container } from '../contented/container';
import { Content } from '../contented/content';

import { ApiDef } from '../contented/api_def';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { GlobalNavEvents, NavTypes } from '../contented/nav_events';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';

describe('TestingContainerNavCmp', () => {
  let fixture: ComponentFixture<ContainerNavCmp>;
  let service: ContentedService;
  let comp: ContainerNavCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;
  let listener: Subscription;
  let cnt: Container;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [ContentedService],
      teardown: { destroyAfterEach: false },
    }).compileComponents();

    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(ContainerNavCmp);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.container-nav-cmp'));
    el = de.nativeElement;
    cnt = new Container(MockData.getPreview()[0]);

    let res = MockData.getContent(cnt.id, 5);
    let contents = _.map(res.results, c => new Content(c));
    cnt.addContents(contents);

    listener = GlobalNavEvents.navEvts.subscribe(evt => {
      if (evt.action == NavTypes.NEXT_MEDIA) {
        GlobalNavEvents.selectContent(cnt.getContent(++cnt.rowIdx), cnt);
      } else if (evt.action == NavTypes.PREV_MEDIA) {
        GlobalNavEvents.selectContent(cnt.getContent(--cnt.rowIdx), cnt);
      }
    });
  }));

  afterEach(() => {
    if (listener) {
      listener.unsubscribe();
    }
  });

  function getRowIdx() {
    const val = $('.cnt-row-idx').val();
    if (typeof val === 'string') {
      return parseInt(val, 10);
    }
    return 0;
  }

  it('Should create a contented component', () => {
    expect(comp).toBeDefined('We should have the Contented comp');
    expect(el).toBeDefined('We should have a top level element');
  });

  it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
    cnt.total = 10;
    comp.cnt = cnt;

    fixture.detectChanges();
    expect($('.container-bar').length).toEqual(1, 'It should have a container bar');
    let fullLoadBtn = $('.btn-full-load-ctn');
    expect(fullLoadBtn.length).toEqual(1, 'We should have a full load btn');
    fullLoadBtn.trigger('click');
    fixture.detectChanges();

    let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id.toString());
    let req = httpMock.expectOne(r => r.url === url);
  }));

  it('Should be able to navigate by button clicks', fakeAsync(() => {
    cnt.total = 10;
    comp.cnt = cnt;
    fixture.detectChanges();

    let nextBtn = $('.nav-content-next');
    let prevBtn = $('.nav-content-previous');
    let rowIdx = $('.cnt-row-idx');

    expect(rowIdx.length).withContext('There should be an input for jumping to content').toEqual(1);
    expect(nextBtn.length).withContext('It should have a next button').toEqual(1);
    expect(prevBtn.length).withContext('It should have a previous button').toEqual(1);

    let rowIdxOriginal = cnt.rowIdx;
    expect(getRowIdx()).withContext('It should be on the first element').toEqual(rowIdxOriginal);
    nextBtn.trigger('click');
    nextBtn.trigger('click');
    fixture.detectChanges();
    tick(100);
    expect(getRowIdx())
      .withContext('Now we should be on the next element')
      .toEqual(rowIdxOriginal + 2);

    prevBtn.trigger('click');
    fixture.detectChanges();
    tick(100);
    expect(getRowIdx())
      .withContext('It should go back one')
      .toEqual(rowIdxOriginal + 1);
  }));
});
