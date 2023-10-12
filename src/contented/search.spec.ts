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

import {SearchCmp} from '../contented/search.cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';

import * as _ from 'lodash';
import {MockData} from '../test/mock/mock_data';

declare var $;
describe('TestingSearchCmp', () => {
    let fixture: ComponentFixture<SearchCmp>;
    let service: ContentedService;
    let comp: SearchCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let router: Router;

    let httpMock: HttpTestingController;
    let loc: Location;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule.withRoutes(
                    [{path: 'ui/search', component: SearchCmp}]
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
        fixture = TestBed.createComponent(SearchCmp);
        httpMock = TestBed.inject(HttpTestingController);
        loc = TestBed.inject(Location);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.search-cmp'));
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
        router.navigate(["/ui/search/"], {queryParams: {searchText: st}}); 
        tick(100);
        console.log("navigate is done");
        fixture.detectChanges();
        let vals = comp.getValues();
        console.log(vals);
        tick(100);
        expect(vals['searchText']).withContext("It should default via route params").toBe(st);

        let req = httpMock.expectOne(req => req.url === ApiDef.contented.search);
        let sr = MockData.getSearch()
        expect(sr.results.length).withContext("We need some search results.").toBeGreaterThan(0);
        req.flush(sr);
        fixture.detectChanges();
        expect($('.search-result').length).withContext("It should render dom results").toEqual(sr.results.length);
    }));
});

