import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {ContentBrowserCmp} from '../contented/content_browser.cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';
import {GlobalNavEvents} from '../contented/nav_events';

import * as _ from 'lodash';
import {MockData} from '../test/mock/mock_data';

declare var $;
describe('TestingContentBrowserCmp', () => {
    let fixture: ComponentFixture<ContentBrowserCmp>;
    let service: ContentedService;
    let comp: ContentBrowserCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/browse/:idx/:rowIdx', component: ContentBrowserCmp}]
                ),
                FormsModule,
                ContentedModule,
                HttpClientTestingModule,
                NoopAnimationsModule,
            ],
            providers: [
                ContentedService
            ],
            teardown: {destroyAfterEach: false},
        }).compileComponents();

        service = TestBed.get(ContentedService);
        fixture = TestBed.createComponent(ContentBrowserCmp);
        httpMock = TestBed.get(HttpTestingController);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.content-browser-cmp'));
        el = de.nativeElement;
        router = TestBed.get(Router);
        router.initialNavigation();
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('TODO: Fully handles routing arguments', fakeAsync(() => {
        // Should just setup other ajax calls
        router.navigate(['/ui/browse/2/3']);
        tick(100);
        expect(router.url).toBe('/ui/browse/2/3');
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        tick(1000);
        // TODO: Make a test that actually works with the damn activated route params
        // expect(comp.idx).toBe(2, "It should pull the dir index from ");
    }));

    function handleContainerContentLoad(dirs: Array<Container>) {
        _.each(dirs, dir => {
            let url = ApiDef.contented.containerContent.replace('{cId}', dir.id);
            let req = httpMock.expectOne(r => r.url === url);
            req.flush(MockData.getContent(dir.name, 2));
        });
    }

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        tick(2000);
        expect(comp.allCnts.length).toBe(4, "We should have 4 containers set");

        let dirs = comp.getVisibleContainers();
        expect(dirs.length).toBe(comp.maxVisible, "Should only have the max visible containers present.");
        expect(dirs.length <= comp.allCnts.length).toBe(true, "It should never have more data than we asked for.");

        fixture.detectChanges();
        let dirEls = $('.container-contents', el);
        expect(dirEls.length).toBe(comp.maxVisible, "We should have the elements rendered.");
        expect($('.current-content-cnt').length).toBe(1, "We should only have 1 selected cnt");
    }));

    it("Should be able to tell you that nothing was loaded up", fakeAsync(() => {
        expect(comp.emptyMessage).toBe(null);
        expect($('.no-content').length).toBe(0, "Nothing is loaded.");
        fixture.detectChanges();

        let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
        containersReq.flush([]);
        tick(1000);
        fixture.detectChanges();

        expect(comp.emptyMessage).not.toBe(null, "Now we should have an error message");
        expect($('.no-content').length).toBe(1, "We should now have a visible UI msg");
    }));


    it('Should handle a click event to show a particular image.', fakeAsync(() => {
        fixture.detectChanges();
        tick(2000);

        let containers = MockData.getPreview();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        expect($('.content-full-view').length).toBe(0, "It should not have a view");
        fixture.detectChanges();
        tick(1000);
        fixture.detectChanges();

        let cnt = comp.getCurrentContainer();
        expect(cnt).toBeDefined("There should be a current container");
        let cl = cnt.getContentList();
        expect(cl).toBeDefined("We should have a content list");
        expect(cl.length).toEqual(2, "And we should have content");

        fixture.detectChanges();
        let imgs = $('.preview-img');
        expect(imgs.length).toBeGreaterThan(2, "A bunch of images should be visible");
        expect($('.content-full-view').length).toBe(0, "It should not have a view");

        let toClick = $(imgs[3]).trigger('click');
        fixture.detectChanges();

        let currLoc = $('.current-img');
        let fullView = $('.full-view-img');
        expect(currLoc.length).toBe(1, "It should be selected still");
        expect(fullView.length).toBe(1, "It should now have a view");
        expect(currLoc.prop('src').replace("preview", "view")).toBe(
            fullView.prop("src"), "The full view should be /view/selectedId"
        );

        // Because we select the 3rd image in the NEXT container set it should
        // now attempt to load data from the newly visible container
        let cnts = comp.getVisibleContainers()
        let nextContainer = cnts[1];
        let url = ApiDef.contented.containerContent.replace("{cId}", nextContainer.id);
        httpMock.expectOne(r => r.url == url);
        tick(1000);
    }));

    it("Should have a progress bar once the data is loaded", () => {
        // Kick off a load and use the http controller mocks to return our containers
        fixture.detectChanges();

        let containers = MockData.getPreview();
        MockData.handleCmpDefaultLoad(httpMock, fixture);

        expect(comp.loading).toBe(false, "It should be fine with loading the containers");
        expect(comp.allCnts.length).toBeGreaterThan(0, "There should be a number of containers");
        fixture.detectChanges();

        expect(comp.idx).toBe(0, "It should be on the default page");
        let dirs = $('.cnt-name');
        expect(dirs.length).toBe(2, "There should be two containers present");
        expect(_.get(containers, '[0].name')).toBe($(dirs[0]).text(), 'It should have the dir id');
        expect(_.get(containers, '[1].name')).toBe($(dirs[1]).text(), 'It should have the dir name');

        let progBars = $('mat-progress-bar');
        expect(progBars.length).toBe(2, "We should have two rendered bars");
        expect($(progBars.get(0)).attr('mode')).toBe('buffer', "First dir is not fully loaded");
    });


    it('Pull in more more contents in a dir', fakeAsync(() => {
        fixture.detectChanges();

        MockData.handleCmpDefaultLoad(httpMock, fixture);
        fixture.detectChanges();
        tick(1000);

        let dir: Container = comp.getCurrentContainer();
        expect(dir).not.toBe(null);
        expect(dir.count).toBeLessThan(dir.total, "There should be more to load");
        expect(dir.count).toEqual(2, "The default count should be 2 by default");

        let prevCount = dir.count;
        expect(prevCount).toBe(2, "The default load has 2 items");

        service.LIMIT = 1;
        comp.loadMore();
        let url = ApiDef.contented.containerContent.replace('{cId}', dir.id);
        let loadReq = httpMock.expectOne(req => req.url === url);
        let checkParams: HttpParams = loadReq.request.params;
        expect(checkParams.get('per_page')).toBe('1', "We set a different limit");


        let page = parseInt(checkParams.get('page'), 10);
        let offset = (page) * service.LIMIT;
        expect(page).toBeGreaterThan(2, "It should load more, not the beginning");
        expect(offset).toEqual(3, "Calculating the offset should be more than the current count");

        let content = MockData.getContent(dir.id, service.LIMIT);
        loadReq.flush(content);
        fixture.detectChanges();

        expect(dir.count).toEqual(3, "Now we should have loaded more based on the limit");
        fixture.detectChanges();
    }));

    it('Pull in more more contents in a dir', fakeAsync(() => {
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        fixture.detectChanges();

        let cnt = comp.allCnts[3];
        let content = cnt.getContent();
        // Check that a content in container 3 is not visible
        expect(comp.allCnts.length).toBeGreaterThan(0, "We should have containers");
        expect(comp.idx).toEqual(0, "We should be at index 0");

        GlobalNavEvents.selectContainer(cnt);
        fixture.detectChanges();
        MockData.handleContainerContentLoad(httpMock, [cnt], 3);
        expect(comp.idx).toEqual(3, "We should now be on the third index")
        tick(1000);
        fixture.detectChanges();
    }));
});

