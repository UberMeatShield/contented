import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

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

    beforeEach(async( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/:idx/:rowIdx', component: ContentedCmp}]
                ),
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
        MockData.mockContentedService(comp._contentedService);

        router.navigate(['/ui/2/3']);
        tick(100);
        expect(router.url).toBe('/ui/2/3');

        fixture.detectChanges();
        tick(1000);
        // TODO: Make a test that actually works with the damn activated route params
        // expect(comp.idx).toBe(2, "It should pull the dir index from ");
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        MockData.mockContentedService(comp._contentedService);
        fixture.detectChanges();
        tick(2000);
        expect(comp.allD.length).toBe(3, "We should have 3 directories set");

        let dirs = comp.getVisibleDirectories();
        expect(dirs.length).toBe(comp.maxVisible, "Should only have the max visible directories present.");
        expect(dirs.length <= comp.allD.length).toBe(true, "It should never have more data than we asked for.");

        fixture.detectChanges();
        let dirEls = $('.directory-contents', el);
        expect(dirEls.length).toBe(comp.maxVisible, "We should have the elements rendered.");

        expect($('.current-content-dir').length).toBe(1, "We should only have 1 selected dir");
    }));


    it('Should handle a click event to show a particular image.', fakeAsync(() => {
        MockData.mockContentedService(comp._contentedService);
        fixture.detectChanges();
        tick(2000);

        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode");
        expect($('.contented-view-cmp').length).toBe(0, "It should now have a view component.");

        fixture.detectChanges();
        let imgs = $('.preview-img');
        expect(imgs.length > 1).toBe(true, "A bunch of images should be visible");
        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode even after everything is loaded");

        let toClick = $(imgs[3]).trigger('click');
        expect(comp.fullScreen).toBe(true, "It should now have a selected item");
        expect(comp.getCurrentLocation()).toBe(imgs[3].src, "It should have the current item as the image");
    }));

    it("Should have a progress bar once the data is loaded", () => {
        // Kick off a load and use the http controller mocks to return our preview
        fixture.detectChanges();

        let preview = MockData.getPreview();
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.preview);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);

        expect(comp.loading).toBe(false, "It should be fine with loading the preview");
        expect(comp.allD.length).toBeGreaterThan(0, "There should be a number of directories");
        fixture.detectChanges();

        expect(comp.idx).toBe(0, "It should be on the default page");
        let dirs = $('.dir-name');
        expect(dirs.length).toBe(2, "There should be two directories present");
        expect(_.get(preview, 'results[0].id')).toBe($(dirs[0]).text(), 'It should have the dir id');
        expect(_.get(preview, 'results[1].id')).toBe($(dirs[1]).text(), 'It should have the dir id');

        let progBars = $('mat-progress-bar');
        expect(progBars.length).toBe(2, "We should have two rendered bars");
        expect($(progBars.get(0)).attr('mode')).toBe('buffer', "First dir is not fully loaded");

        // Fully load, check that it is not longer in buffer mode
        // TODO: Go to the next dir
        // Check the current row index (increase the loaded data), go next, check visibile state
    });


    it('Should be able to load more contents in a dir', () => {
        fixture.detectChanges();
        let preview = MockData.getPreview();
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.preview);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);
        fixture.detectChanges();

        let dir: Directory = comp.getCurrentDir();
        expect(dir).not.toBe(null);
        expect(dir.count).toBeLessThan(dir.total, "There should be more to load");
        let prevCount = dir.count;
        expect(prevCount).not.toBe(0, "It should start with content");

        service.LIMIT = 1;
        comp.loadMore();
        let fullUrl = ApiDef.contented.fulldir.replace('{dir}', dir.id);
        let fullReq = httpMock.expectOne(req => req.url === fullUrl);
        let checkParams: HttpParams = fullReq.request.params;
        expect(checkParams.get('limit')).toBe('1', "We set a different limit");
        expect(checkParams.get('offset')).toBe('' + dir.count, "It should load more, not the beginning");

        let firstLoad = MockData.getMockDir(dir.total - dir.count);
        fullReq.flush(firstLoad);
        expect(dir.count).toBeGreaterThan(prevCount, "We should have added more");
        expect(dir.contents.indexOf(firstLoad.contents[0])).toBeGreaterThan(0, 'It should have added an element');
        expect(dir.count).toBe(dir.total, "It should load all the data");
    });
});

