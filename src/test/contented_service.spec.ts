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
});

