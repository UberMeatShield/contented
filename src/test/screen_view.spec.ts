import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Screen} from '../contented/screen';
import {ScreensCmp} from '../contented/screens_cmp';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';
import {GlobalNavEvents} from '../contented/nav_events';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingScreensCmp', () => {
    let fixture: ComponentFixture<ScreensCmp>;
    let service: ContentedService;
    let comp: ScreensCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'screens/:screenId', component: ScreensCmp}]
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
        fixture = TestBed.createComponent(ScreensCmp);
        httpMock = TestBed.get(HttpTestingController);
        comp = fixture.componentInstance;
        de = fixture.debugElement.query(By.css('.screens-cmp'));
        el = de.nativeElement;
        router = TestBed.get(Router);
        router.initialNavigation();
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('Should create a screens view component', () => {
        let screen = new Screen({id: 'a'});
        expect(screen.url).toBeDefined("It should set the link if possible.");
    });

    it('Should build out a screen view and be able to render', () => {
        expect(el).toBeDefined("We should have built out a component.");
        expect($(".screens-cmp").length).toEqual(1, "The component should exist");
    });

    it('Given a media id it will try and render screens', fakeAsync(() => {
        let mediaId = "uuid-really";
        comp.mediaId = mediaId;
        fixture.detectChanges();
        expect(comp.loading).toBeTrue();

        let url = ApiDef.contented.mediaScreens.replace("{mcID}", mediaId);
        let req = httpMock.expectOne(req => req.url == url);
        let screens = MockData.getScreens();
        expect(screens.length).toBeGreaterThan(0, "We should have screens in the mock data");
        req.flush(screens);
        tick(1000);

        fixture.detectChanges();
        expect(comp.loading).toBeFalse(); // It should no longer be loading
        expect(comp.screens.length).toEqual(screens.length, "We should have assigned screens");
        expect($(".screen-img", el).length).toEqual(screens.length, "There should be screens rendered");
        expect($(".screen", el).length).toEqual(screens.length, "There should be screens rendered");
    }));
});
