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

import {VideoViewCmp} from '../contented/video_view_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingVideoViewCmp', () => {
    let fixture: ComponentFixture<VideoViewCmp>;
    let service: ContentedService;
    let comp: VideoViewCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;
    let loc: Location;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/video', component: VideoViewCmp}]
                ),
                FormsModule,
                ContentedModule,
                HttpClientTestingModule,
                NoopAnimationsModule,
            ],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.inject(ContentedService);
        fixture = TestBed.createComponent(VideoViewCmp);
        httpMock = TestBed.inject(HttpTestingController);
        loc = TestBed.inject(Location);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.video-view-cmp'));
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
        console.log("navigate is done");
        fixture.detectChanges();
        let vals = comp.getValues();
        console.log(vals);
        tick(100);
        expect(vals['videoText']).toBe(st, "It should default via route params");

        let req = httpMock.expectOne(req => req.url === ApiDef.contented.search);
        let sr = MockData.getVideos()

        expect(sr['media'].length).toBeGreaterThan(0, "We need some search results.");
        req.flush(sr);
        fixture.detectChanges();
        expect($('.video-view-card').length).toEqual(sr['media'].length);

        let mcId = sr.media[0].id;
        let screenUrl = ApiDef.contented.mediaScreens.replace("{mcID}", mcId);
        let screenReq = httpMock.expectOne(req => req.url == screenUrl);
        screenReq.flush(MockData.getScreens());
        tick(100);
    }));
});

