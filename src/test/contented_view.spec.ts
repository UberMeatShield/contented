import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {ContentedViewCmp} from '../contented/contented_view_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Media} from '../contented/media';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';
import {GlobalNavEvents} from '../contented/nav_events';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingContentedViewCmp', () => {
    let fixture: ComponentFixture<ContentedViewCmp>;
    let service: ContentedService;
    let comp: ContentedViewCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/:idx/:rowIdx', component: ContentedViewCmp}]
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
        fixture = TestBed.createComponent(ContentedViewCmp);
        httpMock = TestBed.get(HttpTestingController);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-view-cmp'));
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

    it('Can render an image and render', () => {
        comp.media = null;
        comp.visible = true;
        fixture.detectChanges();
        expect($('.media-full-view').length).toBe(0, "It should not be visible");

        let img = MockData.getImg();
        comp.media = img;
        fixture.detectChanges();
        expect($('.media-full-view').length).toBe(1, "It should be visible");
    });

    it('Forcing a width and height will be respected', () => {
        comp.media = MockData.getImg();
        comp.forceWidth = 666;
        comp.forceHeight = 42;
        comp.visible = true;
        fixture.detectChanges();
        expect($('.media-full-view').length).toBe(1, "It should be visible");

        window.dispatchEvent(new Event('resize'));
        fixture.detectChanges();
        // It should be forcing a detection of the resize (otherwise it is calculated already)
        // comp.calculateDimensions();

        expect(comp.maxWidth).toEqual(comp.forceWidth, "Ensure width assignment works");
        expect(comp.maxHeight).toEqual(comp.forceHeight, "Ensure height assignment works");
    });

    // Test that we listen to nav events correctly
    it('Should register nav events', () => {
        fixture.detectChanges();
        expect($('.media-full-view').length).toBe(0, "Nothing in the view");

        let initialSel = new Media({id: 'A'})
        GlobalNavEvents.selectMedia(initialSel, new Container({id: '1'}));
        fixture.detectChanges();
        expect(comp.media).toEqual(initialSel);

        let media = MockData.getImg();
        GlobalNavEvents.viewFullScreen(media);
        fixture.detectChanges();
        expect($('.media-full-view').length).toBe(1, "It should now be visible");
        expect($('.full-view-img').length).toBe(1, "And it is an image");
        expect(comp.media).toEqual(media, "A view event with a media item should change it");

        GlobalNavEvents.hideFullScreen();
        fixture.detectChanges();
        expect(comp.visible).toBe(false, "It should not be visible now");
    });
});
