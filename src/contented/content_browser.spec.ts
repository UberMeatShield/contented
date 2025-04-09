import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement, NgZone } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { RouterTestingHarness, RouterTestingModule } from '@angular/router/testing';
import { provideRouter, Router } from '@angular/router';

import { ContentBrowserCmp } from '../contented/content_browser.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule, WaitForMonacoLoad } from '../contented/contented_module';
import { Container } from '../contented/container';
import { Content } from '../contented/content';
import { ApiDef } from '../contented/api_def';
import { GlobalNavEvents } from '../contented/nav_events';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';

describe('TestingContentBrowserCmp', () => {
  let service: ContentedService;
  let comp: ContentBrowserCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;
  let harness: RouterTestingHarness;

  beforeEach(async () => {
    TestBed.configureTestingModule({
      imports: [FormsModule, ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [ContentedService, provideRouter([{ path: 'ui/browse/:idx/:rowIdx', component: ContentBrowserCmp }])],
      teardown: { destroyAfterEach: true },
    }).compileComponents();

    harness = await RouterTestingHarness.create();
    comp = await harness.navigateByUrl('/ui/browse/0/0', ContentBrowserCmp);
    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    el = harness.fixture.debugElement.nativeElement;
    de = harness.fixture.debugElement;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('TODO: Fully handles routing arguments', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/browse/2/3', ContentBrowserCmp);
    comp.allCnts = [new Container({ id: 2, total: 0, count: 0, contents: [] })];
    harness.detectChanges();
    router = TestBed.inject(Router);
    expect(router.url).toBe('/ui/browse/2/3');
    harness.detectChanges();
    MockData.handleCmpDefaultLoad(httpMock, harness);
  }));

  it('Should create a contented component', async () => {
    harness.detectChanges();
    const el = harness.fixture.debugElement.query(By.css('.content-browser-cmp'));
    expect(comp).toBeDefined('We should have the Contented comp');
    expect(el).toBeDefined('We should have a top level element');
    httpMock.expectOne(ApiDef.contented.containers).flush({ results: [] });
  });

  it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
    harness.detectChanges();
    MockData.handleCmpDefaultLoad(httpMock, harness);
    tick(2000);
    expect(comp.allCnts.length).withContext('We should have 4 containers set').toBe(6);

    let dirs = comp.getVisibleContainers();
    expect(dirs.length).toBe(comp.maxVisible, 'Should only have the max visible containers present.');
    expect(dirs.length <= comp.allCnts.length).toBe(true, 'It should never have more data than we asked for.');

    harness.detectChanges();
    let dirEls = $('.container-contents', el);
    expect(dirEls.length).toBe(comp.maxVisible, 'We should have the elements rendered.');
    expect($('.current-content-cnt').length).toBe(1, 'We should only have 1 selected cnt');
  }));

  it('Should be able to tell you that nothing was loaded up', fakeAsync(() => {
    expect(comp.emptyMessage).toBe(undefined);
    expect($('.no-content').length).toBe(0, 'Nothing is loaded.');
    harness.detectChanges();

    let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
    containersReq.flush([]);
    tick(1000);
    harness.detectChanges();

    expect(comp.emptyMessage).not.toBe(undefined, 'Now we should have an error message');
    expect($('.no-content').length).toBe(1, 'We should now have a visible UI msg');
  }));

  it('Should handle a click event to show a particular image.', fakeAsync(() => {
    harness.detectChanges();
    tick(2000);

    MockData.getPreview();
    MockData.handleCmpDefaultLoad(httpMock, harness);
    expect($('.content-full-view').length).withContext('It should not have a view').toBe(0);
    tick(10000);
    harness.detectChanges();
    tick(10000);
    harness.detectChanges();

    let cnt = comp.getCurrentContainer();
    expect(cnt).toBeDefined();
    if (!cnt) {
      throw new Error("no current test container.")
    }
    expect(cnt).withContext('There should be a current container').toBeDefined();
    const arr = MockData.getContentArr(cnt?.id, cnt?.total);
    cnt?.addContents(arr);

    let cl = cnt?.getContentList() || [];
    expect(cl).withContext('We should have a content list').toBeDefined();
    expect(cl.length).withContext('And we should have content').toEqual(cnt?.total);

    harness.detectChanges();
    let imgs = $('.preview-img');
    expect(imgs.length).withContext('A bunch of images should be visible').toBeGreaterThan(2);
    expect($('.content-full-view').length).withContext('It should not have a view').toBe(0);

    let toClick = $(imgs[2]).trigger('click');
    expect(toClick).toBeDefined();
    tick(100);
    harness.detectChanges();

    let currLoc = $('img', $('.current-content'));
    let fullView = $('.full-view-img');
    expect(currLoc.length).withContext('It should be selected still').toBe(1);
    expect(fullView.length).withContext('It should now have a view').toBe(1);
    expect(currLoc.prop('src').replace('preview', 'view')).toBe(
      fullView.prop('src'),
      'The full view should be /view/selectedId'
    );
    tick(1000);
    harness.detectChanges();
  }));

  it('Should have a progress bar once the data is loaded', () => {
    // Kick off a load and use the http controller mocks to return our containers
    harness.detectChanges();

    let cntRes = MockData.getPreview();
    let containers = cntRes.results;
    MockData.handleCmpDefaultLoad(httpMock, harness);

    expect(comp.loading).toBe(false, 'It should be fine with loading the containers');
    expect(comp.allCnts.length).toBeGreaterThan(0, 'There should be a number of containers');
    harness.detectChanges();

    expect(comp.idx).toBe(0, 'It should be on the default page');
    let dirs = $('.cnt-name');
    expect(dirs.length).toBe(2, 'There should be two containers present');
    expect(_.get(containers, '[0].name')).toBe($(dirs[0]).text(), 'It should have the dir id');
    expect(_.get(containers, '[1].name')).toBe($(dirs[1]).text(), 'It should have the dir name');

    let progBars = $('mat-progress-bar');
    expect(progBars.length).toBe(2, 'We should have two rendered bars');

    expect($(progBars[0]).attr('mode')).toBe('buffer', 'First dir is not fully loaded');
  });

  it('Pull in more contents in a dir', fakeAsync(() => {
    harness.detectChanges();
    MockData.handleCmpDefaultLoad(httpMock, harness);
    harness.detectChanges();
    tick(1000);

    let cnt: Container | undefined = comp.getCurrentContainer();
    expect(cnt).not.toBe(undefined);
    if (!cnt) {
      throw new Error("no current test container.")
    }
    cnt.contents = [];
    cnt.total = 4;
    cnt.count = 0;

    expect(cnt.total).withContext('There should be more to load').toBeGreaterThan(3);
    expect(cnt.count).withContext('The default count should be empty').toEqual(0);
    cnt.addContents(MockData.getContentArr(cnt.id, 2));
    expect(cnt.count).withContext('Added some default data').toEqual(2);

    service.LIMIT = 1;
    comp.loadMore();
    let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id.toString());
    let loadReq = httpMock.expectOne(req => req.url === url);
    let checkParams: HttpParams = loadReq.request.params;
    expect(checkParams.get('per_page')).withContext('We set a different limit').toBe('1');

    let page = parseInt(checkParams.get('page') || '0', 10);
    let offset = page * service.LIMIT;
    expect(page).withContext('It should load more, not the beginning').toBeGreaterThan(0);
    expect(offset).withContext('Calculating the offset should be more than the current count').toEqual(3);
    tick(100);
    harness.detectChanges();

    const contentsAll = MockData.getContent(cnt.id, cnt.total);
    const results = contentsAll.results;
    contentsAll.results = results.slice(cnt.count, cnt.count + 1);

    loadReq.flush(contentsAll);
    harness.detectChanges();
    tick(100);
    expect(cnt.count).withContext('Now we should have loaded more based on the limit').toEqual(3);
    harness.detectChanges();
  }));

  it('Ensure indexing works at least somewhat and loads the last selected', fakeAsync(() => {
    harness.detectChanges();
    MockData.handleCmpDefaultLoad(httpMock, harness);
    harness.detectChanges();
    tick(10000); // Important to let the paged loading finish

    expect(_.isEmpty(comp.allCnts)).withContext('We should have content').toBeFalse();
    expect(comp.allCnts.length).withContext('We should have containers').toBeGreaterThan(4);

    let lastIdx = comp.allCnts.length - 1;
    let cnt = comp.allCnts[lastIdx];
    expect(comp.idx).withContext('We should be at index 0').toEqual(0);

    console.log('Attempting to select', cnt.id);
    GlobalNavEvents.selectContainer(cnt);
    tick(1000);
    harness.detectChanges();
    tick(10000);
    expect(comp.idx).withContext('We should now be on the last index').toEqual(lastIdx);
    //console.log("Current", comp.getCurrentContainer(), cnt.id);
    harness.detectChanges();
    MockData.handleContainerContentLoad(httpMock, [cnt], 3);
    tick(1000);
    harness.detectChanges();
  }));

  it('Can handle rendering a text element into the page', waitForAsync(async () => {
    let containerId = 3;
    let container = new Container({
      id: containerId,
      total: 1,
      count: 0,
      contents: null,
    });

    let contentId = 42;
    let content = {
      id: contentId,
      content_type: 'text/plain; charset=utf-8',
      container_id: containerId,
      src: '/ab',
    };

    comp = await harness.navigateByUrl('/ui/browse/0/1', ContentBrowserCmp);
    let checkContent = new Content(content);
    expect(checkContent.shouldUseTypedPreview()).toEqual('article');

    harness.detectChanges();
    const cntReq = httpMock.expectOne(ApiDef.contented.containers);
    cntReq.flush({ results: [container] });
    harness.detectChanges();
    expect(comp.allCnts.length).toEqual(1);
    expect(comp.getVisibleContainers().length).toEqual(1);

    let url = ApiDef.contented.containerContent.replace('{cId}', containerId.toString());
    const contentReq = httpMock.match(r => r.url.includes(url));
    contentReq.forEach(r => r.flush({ results: [content] }));
    harness.detectChanges();
    expect($('.contented-cnt').length).withContext('We should have a container').toEqual(1);
    harness.detectChanges();

    harness.detectChanges();
    let contentDom = $('.preview-content');
    expect(contentDom.length).withContext("We don't have some sort of item").toEqual(1);

    /*  TODO: Currently the monaco editor does NOT play nice with the new route testing harness.
    contentDom.trigger('click');
    harness.detectChanges();
    await harness.fixture.whenRenderingDone();
    harness.detectChanges();

    await WaitForMonacoLoad();

    httpMock.expectOne(ApiDef.contented.download.replace('{mcID}', contentId)).flush('What');
    harness.detectChanges();
    expect($('.preview-type').length).withContext('There should be a text editor').toEqual(1);
    harness.detectChanges();

    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(MockData.tags());
    expect($('vscode-editor-cmp').length).withContext('Text should load the editor').toEqual(1);
    */
  }));
});
