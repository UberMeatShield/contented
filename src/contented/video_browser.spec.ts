import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';

import {Location} from '@angular/common';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {VideoBrowserCmp} from '../contented/video_browser.cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';

import * as _ from 'lodash';
import {MockData} from '../test/mock/mock_data';

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

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/video', component: VideoBrowserCmp}]
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

        service = TestBed.inject(ContentedService);
        fixture = TestBed.createComponent(VideoBrowserCmp);
        httpMock = TestBed.inject(HttpTestingController);
        loc = TestBed.inject(Location);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.video-browser-cmp'));
        el = de.nativeElement;
        router = TestBed.get(Router);
        router.initialNavigation();
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('It can setup all eventing without exploding', fakeAsync(() => {
        let st = "Cthulhu";
        router.navigate(["/ui/video/"], {queryParams: {videoText: st}}); 
        tick(100);
        fixture.detectChanges();
        let vals = comp.getValues();
        tick(100);
        expect(vals['videoText']).toBe(st, "It should default via route params");

        MockData.handleContainerLoad(httpMock);

        let req = httpMock.expectOne(req => req.url === ApiDef.contented.search);
        let sr = MockData.getVideos()

        expect(sr['content'].length).toBeGreaterThan(0, "We need some search results.");
        req.flush(sr);
        fixture.detectChanges();
        expect($('.video-view-card').length).toEqual(sr['content'].length);
        tick(100);
    }));

    it("Will load up screens if they are not provided", fakeAsync(() => {
        let vRes = MockData.getVideos()
        _.each(vRes.content, v => {
            v.screens = null;
        });

        fixture.detectChanges();
        MockData.handleContainerLoad(httpMock);
        let req = httpMock.expectOne(req => req.url === ApiDef.contented.search);
        req.flush(vRes);
        fixture.detectChanges();
        tick(100);

        _.each(vRes.content, mc => {
            let screenUrl = ApiDef.contented.contentScreens.replace("{mcID}", mc.id);
            let screenReq = httpMock.expectOne(req => req.url == screenUrl);
            screenReq.flush(MockData.getScreens());
        });
        tick(1000);
        fixture.detectChanges();
        expect($(".video-details").length).toEqual(4, "Should show 4 details sections");
    }));
});
