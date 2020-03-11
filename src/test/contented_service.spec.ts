import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpParams, HttpRequest} from '@angular/common/http';
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

describe('TestingContentedService', () => {
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
        service.getPreview().subscribe(
            (dirs: Array<Directory>) => {
                expect(dirs.length).toEqual(preview['results'].length, "It should kick back data");

                _.each(dirs, dir => {
                    expect(dir.count).toBeGreaterThan(0, "All of them should have contents");
                    expect(dir.count).toBe(dir.contents.length, "It should equal out");
                });
                reallyRan = true;
            },
            err => {
                fail(err);
            }
        );
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.preview);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);
        expect(reallyRan).toBe(true);
    });


    it('Can load up a lot of data and get the offset right', fakeAsync(() => {
        const total = 50001;
        let response = MockData.getMockDir(10000, 'i-', 0, total);
        let dir = new Directory(response);

        expect(dir.contents.length).toEqual(10000, 'Ensure the tests generates correclty');
        expect(dir.id).toBeTruthy();

        service.fullLoadDir(dir, 5000);

        let url = ApiDef.contented.fulldir.replace('{dir}', dir.id);
        let calls = httpMock.match((req: HttpRequest<any>) => {
            return req.url === url;
        });
        expect(calls.length).toEqual(9, 'It should make a lot of calls');

        let expectedMaxFound = false;
        _.each(calls, req => {
            const limit = parseInt(req.request.params.get('limit'), 10);
            const offset = parseInt(req.request.params.get('offset'), 10);

            expect(limit).toEqual(5000, 'The limit should be adjusted');
            expect(offset).toBeGreaterThan(9999, 'The offset should be increasing');
            if (offset === 50000) {
               expectedMaxFound = true;
            }
            const toCreate = offset + limit < total ? limit : total - offset;
            req.flush(MockData.getMockDir(toCreate, 'i-', offset, total));
        });
        tick(1000);
        expect(dir.contents.length).toEqual(total, 'It should load all content');
        expect(dir.contents[total - 1].fullPath).toBeTruthy();

        httpMock.verify();
    }));

    it('Can load the entire directory', fakeAsync(() => {
        let dirs: Array<Directory> = null;
        service.getPreview().subscribe(
            (previewDirs: Array<Directory>) => {
                dirs = previewDirs;
            },
            err => {
                fail(err);
            }
        );
        let preview = _.clone(MockData.getPreview());

        let total = 30;
        preview['results'][0].total = total;
        let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.preview);
        let params: HttpParams = previewReq.request.params;
        previewReq.flush(preview);

        expect(dirs.length).toBeGreaterThan(1, "Should have directories");
        expect(dirs[0].total).toEqual(total);

        let fullDir = dirs[0];
        expect(fullDir.count).toBeLessThan(fullDir.total, "We should not be loaded");

        let loaded: Directory;
        let expectedNumberCalls = fullDir.total - fullDir.count;
        service.fullLoadDir(fullDir, 1).subscribe(
            (dir: Directory) => {
                loaded = dir;
            }, err => { fail(err); }
        );
        let url = ApiDef.contented.fulldir.replace('{dir}', fullDir.id);
        let calls = httpMock.match((req: HttpRequest<any>) => {
            return req.url === url;
        });
        expect(calls.length).toBe(expectedNumberCalls, "We should continue to load");
    }));
});

