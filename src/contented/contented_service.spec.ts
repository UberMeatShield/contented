import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams, HttpRequest } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { Container } from '../contented/container';
import { ApiDef } from '../contented/api_def';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';

describe('TestingContentedService', () => {
  let service: ContentedService;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [ContentedModule, HttpClientTestingModule],
      providers: [ContentedService],
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
      res => {
        expect(res.total).withContext('It should have a count').toEqual(preview.results.length);

        let cnts = res.results;
        _.each(cnts, cnt => {
          expect(cnt.name).withContext('They should have names').toBeDefined();
          expect(cnt.total).withContext('A Total should exist').toBeGreaterThan(0);
          expect(cnt.count).withContext('And data should be loaded').toBe(0);
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

  it('Will be able to get some pagination params correct', () => {
    let params = service.getPaginationParams(1001, 100);
    expect(params.get('page')).toEqual('11', 'We should be on page 11');
    expect(params.get('per_page')).toEqual('100', 'The limits should not change');
  });

  it('Can load up a lot of data and get the offset right', fakeAsync(() => {
    const total = 50001;
    let response = MockData.getMockDir(10000, 'i-', 0, total);
    let dir = new Container(response);

    expect(dir.contents.length).toEqual(10000, 'Ensure the tests generates correclty');
    expect(dir.id).toBeTruthy();
    expect(dir.count).toEqual(dir.contents.length, 'It should have the right count');

    service.fullLoadDir(dir, 5000);

    let url = ApiDef.contented.containerContent.replace('{cId}', dir.id.toString());
    let calls = httpMock.match(r => r.url.includes(url));
    expect(calls.length).toEqual(9, 'It should make a lot of calls');

    let expectedMaxFound = false;
    _.each(calls, req => {
      const limit = parseInt(req.request.params.get('per_page'), 10);
      const page = parseInt(req.request.params.get('page'), 10);
      expect(page).withContext('There should only be 9 page requests, we already loaded 2').toBeLessThan(12);
      const offset = (page - 1) * limit;

      expect(limit).toEqual(5000, 'The limit should be adjusted');
      expect(offset).toBeGreaterThan(9999, 'The offset should be increasing');
      if (page) {
        expectedMaxFound = true;
      }
      const toCreate = offset + limit < total ? limit : total - offset;

      let resN = MockData.getContentsResponse(toCreate, 'i-', offset, total);
      expect(resN.results.length).toEqual(toCreate, 'It should create N entries');
      req.flush(resN);
    });
    tick(100000);
    expect(dir.contents.length).toEqual(total, 'It should load all content');
    expect(expectedMaxFound).toEqual(true, 'It should ');

    httpMock.verify();
  }));

  it('Can load the entire container', fakeAsync(() => {
    let cnts: Array<Container> = null;
    service.getContainers().subscribe(
      res => {
        cnts = res.results;
      },
      err => {
        fail(err);
      }
    );
    let cntRes = _.clone(MockData.getPreview());

    let total = 30;
    cntRes.results[0].total = total;
    let previewReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
    let params: HttpParams = previewReq.request.params;
    previewReq.flush(cntRes);

    expect(cnts.length).toBeGreaterThan(1, 'Should have containers');
    expect(cnts[0].total).toEqual(total);

    let content = cnts[0];
    expect(content.count).toBeLessThan(content.total, 'We should not be loaded');

    let loaded: Container;
    let expectedNumberCalls = content.total - content.count;
    service.fullLoadDir(content, 1).subscribe(
      (dir: Container) => {
        loaded = dir;
      },
      err => {
        fail(err);
      }
    );
    let url = ApiDef.contented.containerContent.replace('{cId}', content.id.toString());
    let calls = httpMock.match((req: HttpRequest<any>) => {
      return req.url === url;
    });
    expect(calls.length).toBe(expectedNumberCalls, 'We should continue to load');
  }));
});
