import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';

import { RouterTestingHarness } from '@angular/router/testing';
import { AdminContainersCmp } from '../contented/admin_containers.cmp';
import { Container } from '../contented/container';

import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';

import { MockData } from '../test/mock/mock_data';
import { ApiDef } from './api_def';
import { provideRouter, Router } from '@angular/router';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import $ from 'jquery';
import { describe } from 'vitest';

describe('TestingAdminContainersCmp', () => {
  let fixture: ComponentFixture<AdminContainersCmp>;
  let service: ContentedService;
  let comp: AdminContainersCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;
  let harness: RouterTestingHarness;

  beforeEach(waitForAsync(async () => {
    TestBed.configureTestingModule({
      imports: [ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [ContentedService, provideRouter([{ path: 'ui/admin/containers', component: AdminContainersCmp }])],
      teardown: { destroyAfterEach: false },
    }).compileComponents();

    harness = await RouterTestingHarness.create();

    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/admin/containers', AdminContainersCmp);
    de = harness.fixture.debugElement.query(By.css('.admin-containers-cmp'));
    el = de.nativeElement;
    expect(comp).withContext('We should have the Contented comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
    harness.detectChanges();

    const req = httpMock.expectOne(r => r.url.includes(ApiDef.contented.searchContainers));
    req.flush({ results: [] });
  }));

  it('It can load up some containers', waitForAsync(async () => {
    comp = await harness.navigateByUrl('/ui/admin/containers', AdminContainersCmp);
    harness.detectChanges();

    const containers = MockData.getContainers();
    const req = httpMock.expectOne(r => r.url.includes(ApiDef.contented.searchContainers));
    req.flush(containers);
    expect(containers?.results?.length).toBeGreaterThan(0);

    //const tagReq = httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags));
    //tagReq.flush(MockData.tags());

    harness.detectChanges();
    expect($('.admin-cnt').length).toEqual(containers.results.length);
  }));
});
