import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
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
import {Directory} from '../contented/directory';
import {ApiDef} from '../contented/api_def';

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

    beforeEach(async( () => {
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
        fixture.detectChanges();
        expect($('.container-full-view').length).toBe(0, "It should not be visible");

        let img = MockData.getImg();
        comp.container = img;
        fixture.detectChanges();
        expect($('.container-full-view').length).toBe(1, "It should not be visible");
    });
});

