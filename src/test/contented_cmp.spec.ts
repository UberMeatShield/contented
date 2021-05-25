import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {ContentedCmp} from '../contented/contented_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Directory} from '../contented/directory';
import {ApiDef} from '../contented/api_def';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingContentedCmp', () => {
    let fixture: ComponentFixture<ContentedCmp>;
    let service: ContentedService;
    let comp: ContentedCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/:idx/:rowIdx', component: ContentedCmp}]
                ),
                FormsModule,
                ContentedModule,
                HttpClientTestingModule
            ],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.get(ContentedService);
        fixture = TestBed.createComponent(ContentedCmp);
        httpMock = TestBed.get(HttpTestingController);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-cmp'));
        el = de.nativeElement;
        router = TestBed.get(Router);
        router.initialNavigation();
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('TODO: Fully handles routing arguments', fakeAsync(() => {
        // Should just setup other ajax calls
        router.navigate(['/ui/2/3']);
        tick(100);
        expect(router.url).toBe('/ui/2/3');
        fixture.detectChanges();
        MockData.handleCmpDefaultLoad(httpMock, fixture);
        tick(1000);
        // TODO: Make a test that actually works with the damn activated route params
        // expect(comp.idx).toBe(2, "It should pull the dir index from ");
    }));

    function handleContainerMediaLoad(dirs: Array<Directory>) {
        _.each(dirs, dir => {
            let url = ApiDef.contented.media.replace('{dirId}', dir.id);
            let req = httpMock.expectOne(r => r.url === url);
            req.flush(MockData.getMedia(dir.dir, 2));
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
        expect(comp.allD.length).toBe(4, "We should have 4 directories set");

        let dirs = comp.getVisibleDirectories();
        expect(dirs.length).toBe(comp.maxVisible, "Should only have the max visible directories present.");
        expect(dirs.length <= comp.allD.length).toBe(true, "It should never have more data than we asked for.");

        fixture.detectChanges();
        let dirEls = $('.directory-contents', el);
        expect(dirEls.length).toBe(comp.maxVisible, "We should have the elements rendered.");

        expect($('.current-content-dir').length).toBe(1, "We should only have 1 selected dir");
    }));


    it('Should handle a click event to show a particular image.', fakeAsync(() => {
        fixture.detectChanges();
        tick(2000);

        let containers = MockData.getPreview();
        MockData.handleCmpDefaultLoad(httpMock, fixture);

        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode");
        expect($('.contented-view-cmp').length).toBe(0, "It should now have a view component.");
        fixture.detectChanges();
        tick(1000);

        let d = comp.getCurrentDir();
        expect(d).toBeDefined("There should be a current directory");
        let cl = d.getContentList();
        expect(cl).toBeDefined("We should have a content list");
        expect(cl.length).toBeGreaterThan(0, "And we should have content");

        let currLoc = comp.getCurrentLocation();
        expect(currLoc).toBeDefined("It should have a current location set");

        fixture.detectChanges();
        let imgs = $('.preview-img');
        expect(imgs.length > 1).toBe(true, "A bunch of images should be visible");
        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode even after everything is loaded");

        let toClick = $(imgs[3]).trigger('click');
        currLoc = comp.getCurrentLocation();
        expect(comp.fullScreen).toBe(true, "It should now have a selected item");
        expect(currLoc.fullUrl).toBe(
            imgs[3].src.replace("preview", "view"),
            "It should have the current item as the image but view"
        );
    }));

    it("Should have a progress bar once the data is loaded", () => {
        // Kick off a load and use the http controller mocks to return our containers
        fixture.detectChanges();

        let containers = MockData.getPreview();
        MockData.handleCmpDefaultLoad(httpMock, fixture);

        expect(comp.loading).toBe(false, "It should be fine with loading the containers");
        expect(comp.allD.length).toBeGreaterThan(0, "There should be a number of directories");
        fixture.detectChanges();

        expect(comp.idx).toBe(0, "It should be on the default page");
        let dirs = $('.dir-name');
        expect(dirs.length).toBe(2, "There should be two directories present");
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

        let dir: Directory = comp.getCurrentDir();
        expect(dir).not.toBe(null);
        expect(dir.count).toBeLessThan(dir.total, "There should be more to load");
        expect(dir.count).toEqual(2, "The default count should be 2 by default");

        let prevCount = dir.count;
        expect(prevCount).toBe(2, "The default load has 2 items");

        service.LIMIT = 1;
        comp.loadMore();
        let url = ApiDef.contented.media.replace('{dirId}', dir.id);
        let loadReq = httpMock.expectOne(req => req.url === url);
        let checkParams: HttpParams = loadReq.request.params;
        expect(checkParams.get('per_page')).toBe('1', "We set a different limit");


        let page = parseInt(checkParams.get('page'), 10);
        let offset = (page) * service.LIMIT;
        expect(page).toBeGreaterThan(2, "It should load more, not the beginning");
        expect(offset).toEqual(3, "Calculating the offset should be more than the current count");

        let media = MockData.getMedia(dir.id, service.LIMIT);
        loadReq.flush(media);
        fixture.detectChanges();

        expect(dir.count).toEqual(3, "Now we should have loaded more based on the limit");
    }));
});

