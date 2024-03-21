import { fakeAsync, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';
import {FormsModule} from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {TagsCmp} from '../contented/tags.cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {ApiDef} from '../contented/api_def';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';

import * as _ from 'lodash';
import {MockData} from '../test/mock/mock_data';

declare var $;

describe('Testing TagsCmp', () => {
    let fixture: ComponentFixture<TagsCmp>;
    let service: ContentedService;
    let comp: TagsCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/view/:id', component: TagsCmp}]
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
        fixture = TestBed.createComponent(TagsCmp);
        httpMock = TestBed.inject(HttpTestingController);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.tags-cmp'));
        el = de.nativeElement;
        router = TestBed.get(Router);
        router.initialNavigation();
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('Should create a contented component', () => {
        expect(comp).withContext("We should have the TagsCmp").toBeDefined();
        expect(el).withContext("We should have a top level element").toBeDefined();
    });

    it('Should be able to render or handle the tags in some way', fakeAsync(() => {
        comp.loadTags = true;
        fixture.detectChanges();

        const req = httpMock.expectOne(r => r.url.includes('/tags/'));
        req.flush(MockData.tags());
        tick(1000);

        expect(comp.tags?.length).toBeGreaterThan(10);
        fixture.detectChanges();
        tick(10000);
    }));
});

