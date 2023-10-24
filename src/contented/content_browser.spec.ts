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
import {Content} from '../contented/content';
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
        expect(comp.allCnts.length).toBe(5, "We should have 4 containers set");

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
        expect($('.content-full-view').length).withContext("It should not have a view").toBe(0);
        tick(10000);
        fixture.detectChanges();
        tick(10000);
        fixture.detectChanges();

        let cnt = comp.getCurrentContainer();
        console.log("What is the current container?", cnt.id)
        expect(cnt).withContext("There should be a current container").toBeDefined()
        cnt.addContents(MockData.getContentArr(cnt.id, 4));
        let cl = cnt.getContentList();
        expect(cl).withContext("We should have a content list").toBeDefined();
        expect(cl.length).withContext("And we should have content").toEqual(4);

        fixture.detectChanges();
        let imgs = $('.preview-img');
        expect(imgs.length).withContext("A bunch of images should be visible").toBeGreaterThan(2);
        expect($('.content-full-view').length).withContext("It should not have a view").toBe(0);

        let toClick = $(imgs[3]).trigger('click');
        fixture.detectChanges();

        let currLoc = $('img', $('.current-content'));
        let fullView = $('.full-view-img');
        expect(currLoc.length).withContext("It should be selected still").toBe(1);
        expect(fullView.length).withContext("It should now have a view").toBe(1);
        expect(currLoc.prop('src').replace("preview", "view")).toBe(
            fullView.prop("src"), "The full view should be /view/selectedId"
        );
        tick(1000);
        fixture.detectChanges();
    }));

    it("Should have a progress bar once the data is loaded", () => {
        // Kick off a load and use the http controller mocks to return our containers
        fixture.detectChanges();

        let cntRes = MockData.getPreview();
        let containers = cntRes.results;
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


    it('Pull in more contents in a dir', fakeAsync(() => {
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        fixture.detectChanges();
        tick(1000);

        let cnt: Container = comp.getCurrentContainer();
        expect(cnt).not.toBe(null);
        expect(cnt.total).withContext("There should be more to load").toBeGreaterThan(3)
        expect(cnt.count).withContext("The default count should be empty").toEqual(0)
        cnt.addContents(MockData.getContentArr(cnt.id, 2));
        expect(cnt.count).withContext("Added some default data").toEqual(2)

        service.LIMIT = 1;
        comp.loadMore();
        let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id);
        let loadReq = httpMock.expectOne(req => req.url === url);
        let checkParams: HttpParams = loadReq.request.params;
        expect(checkParams.get('per_page')).withContext("We set a different limit").toBe('1');


        let page = parseInt(checkParams.get('page'), 10);
        let offset = (page) * service.LIMIT;
        expect(page).withContext("It should load more, not the beginning").toBeGreaterThan(0)
        expect(offset).withContext("Calculating the offset should be more than the current count").toEqual(3);

        let content = MockData.getContent(cnt.id, service.LIMIT);
        loadReq.flush(content);
        fixture.detectChanges();

        expect(cnt.count).withContext("Now we should have loaded more based on the limit").toEqual(3);
        fixture.detectChanges();
    }));

    it('Ensure indexing works at least somewhat and loads the last selected', fakeAsync(() => {
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        fixture.detectChanges();
        tick(10000);  // Important to let the paged loading finish

        expect(_.isEmpty(comp.allCnts)).withContext("We should have content").toBeFalse();
        expect(comp.allCnts.length).withContext("We should have containers").toBeGreaterThan(4);

        let lastIdx = comp.allCnts.length - 1;
        let cnt = comp.allCnts[lastIdx];
        expect(comp.idx).withContext("We should be at index 0").toEqual(0);

        console.log("Attempting to select", cnt.id);
        GlobalNavEvents.selectContainer(cnt);
        tick(1000);
        fixture.detectChanges();
        tick(10000);
        expect(comp.idx).withContext("We should now be on the last index").toEqual(lastIdx);
        //console.log("Current", comp.getCurrentContainer(), cnt.id);
        fixture.detectChanges();
        MockData.handleContainerContentLoad(httpMock, [cnt], 3);
        tick(1000);
        fixture.detectChanges();
    }));

    it("Can handle rendering a text element into the page", fakeAsync(() => {
        let containerId = "A";
        let container = new Container({id: containerId, total: 1, count: 0, contents: null});

        let contentId = "textId";
        let content = {
            id: contentId,
            content_type: "text/plain; charset=utf-8",
            container_id: containerId,
            src: "/ab"
        };
        comp.allCnts = [container];

        let checkContent = new Content(content);
        expect(checkContent.shouldUseTypedPreview()).toEqual("article")
        fixture.detectChanges();

        let url = ApiDef.contented.containerContent.replace("{cId}", containerId);
        httpMock.expectOne(r => r.url.includes(url)).flush({results: [content]});
        expect($(".contented-cnt").length).withContext("We should have a container").toEqual(1);
        expect(comp.allCnts.length).toEqual(1);
        fixture.detectChanges();

        let contentDom = $(".preview-content");
        expect(contentDom.length).withContext("We don't have some sort of item").toEqual(1);
        contentDom.trigger("click");
        tick(1000);

        // Now prove the text is downloaded and added to the editor.
        httpMock.expectOne(ApiDef.contented.download.replace("{mcID}", contentId)).flush("What");
        fixture.detectChanges();
        expect($(".preview-type").length).withContext("There should be a text editor").toEqual(1);
        fixture.detectChanges();
    
        httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(MockData.tags());
        expect($("vscode-editor-cmp").length).withContext("Text should load the editor").toEqual(1);
        tick(10000);
    }));
});