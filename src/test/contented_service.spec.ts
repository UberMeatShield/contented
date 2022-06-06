import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams, HttpRequest} from '@angular/common/http';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import {ContentedCmp} from '../contented/contented_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';
import {Container} from '../contented/container';
import {ApiDef} from '../contented/api_def';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

describe('TestingContentedService', () => {
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
                ContentedModule,
                HttpClientTestingModule
            ],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.get(ContentedService);
        httpMock = TestBed.get(HttpTestingController);
    }));

    afterEach(() => {
        httpMock.verify();
    });

    it('We should be able to load preview data', () => {
        let reallyRan = false;

        let preview = MockData.getPreview();
        service.getContainers().subscribe(
            (dirs: Array<Container>) => {
                expect(dirs.length).toEqual(preview.length, "It should kick back data");

                _.each(dirs, dir => {
                    expect(dir.name).toBeDefined("It should have a name");
                    expect(dir.total).toBeGreaterThan(0, "There should be a total");
                    expect(dir.count).toBe(0, "We have not loaded data at this point");
                });
                reallyRan = true;
            },
            err => {
                fail(err);
            }
        );
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);
        expect(reallyRan).toBe(true);
    });

    it("Will be able to get some pagination params correct", () => {
        let params = service.getPaginationParams(1001, 100);
        expect(params.get("page")).toEqual("11", "We should be on page 11");
        expect(params.get("per_page")).toEqual("100", "The limits should not change");
    });


    it('Can load up a lot of data and get the offset right', fakeAsync(() => {
        const total = 50001;
        let response = MockData.getMockDir(10000, 'i-', 0, total);
        let dir = new Container(response);

        expect(dir.contents.length).toEqual(10000, 'Ensure the tests generates correclty');
        expect(dir.id).toBeTruthy();
        expect(dir.count).toEqual(dir.contents.length, "It should have the right count");

        service.fullLoadDir(dir, 5000);

        let url = ApiDef.contented.media.replace('{cId}', dir.id);
        let calls = httpMock.match((req: HttpRequest<any>) => {
            return req.url === url;
        });
        expect(calls.length).toEqual(9, 'It should make a lot of calls');

        let expectedMaxFound = false;
        _.each(calls, req => {
            const limit = parseInt(req.request.params.get('per_page'), 10);
            const page = parseInt(req.request.params.get('page'), 10);
            expect(page).toBeLessThan(12, "There should only be 9 page requests, we already loaded 2");
            const offset = (page - 1) * limit;

            expect(limit).toEqual(5000, 'The limit should be adjusted');
            expect(offset).toBeGreaterThan(9999, 'The offset should be increasing');
            if (page ) {
               expectedMaxFound = true;
            }
            const toCreate = (offset + limit) < total ? limit : (total - offset);

            let resN = MockData.getMockDir(toCreate, 'i-', offset, total);
            expect(resN.contents.length).toEqual(toCreate, "It should create N entries");
            req.flush(resN.contents);
        });
        tick(100000);
        expect(dir.contents.length).toEqual(total, 'It should load all content');
        expect(expectedMaxFound).toEqual(true, "It should ");

        httpMock.verify();
    }));

    it('Can load the entire container', fakeAsync(() => {
        let dirs: Array<Container> = null;
        service.getContainers().subscribe(
            (previewDirs: Array<Container>) => {
                dirs = previewDirs;
            },
            err => {
                fail(err);
            }
        );
        let preview = _.clone(MockData.getPreview());

        let total = 30;
        preview[0].total = total;
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);

        expect(dirs.length).toBeGreaterThan(1, "Should have containers");
        expect(dirs[0].total).toEqual(total);

        let media = dirs[0];
        expect(media.count).toBeLessThan(media.total, "We should not be loaded");

        let loaded: Container;
        let expectedNumberCalls = media.total - media.count;
        service.fullLoadDir(media, 1).subscribe(
            (dir: Container) => {
                loaded = dir;
            }, err => { fail(err); }
        );
        let url = ApiDef.contented.media.replace('{cId}', media.id);
        let calls = httpMock.match((req: HttpRequest<any>) => {
            return req.url === url;
        });
        expect(calls.length).toBe(expectedNumberCalls, "We should continue to load");
    }));
});

